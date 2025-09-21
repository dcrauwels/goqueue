package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ApiConfig struct {
	DB                   *database.Queries
	Secret               string
	Env                  string
	AccessTokenDuration  int
	RefreshTokenDuration int
}

func (cfg *ApiConfig) GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error) {
	return cfg.DB.GetUserByID(ctx, id)
}

func (cfg ApiConfig) GetSecret() string {
	return cfg.Secret
}

func (cfg ApiConfig) GetEnv() string {
	return cfg.Env
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	cfg.HandlerPostUsers(w, r)

}

func (cfg *ApiConfig) AuthUserMiddleware(next http.Handler) http.Handler {
	/*
		Middleware for user authentication. Note: user is meant in the sense of a /api/users return value here: an employee with an account to summon visitors.
		Returns a handler.	As this returns a handler, no errors are returned. Instead, they are printed to stdout and
		sent to the writer as a response with corresponding HTTP status code.
	*/
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. check for access token cookie
		const usertype string = "user"
		accessTokenCookie, accessErr := r.Cookie("user_access_token")
		if accessErr != nil { // implicitly this means the error returned is http.ErrNoCookie, as request.Cookie can only return that specific error
			// 1.1 if no access token cookie, check for refresh token cookie
			refreshTokenCookie, refreshErr := r.Cookie("user_refresh_token")
			// 1.2 if refresh token cookie present: reissue access token cookie
			if refreshErr == nil { // if accessErr != nil && refreshErr == nil (both errors are guaranteed to be http.ErrNoCookie so this just means
				// there is no cookie with the "access_token" name but there is a "refresh_token" cookie)

				// 1.3 rotate refresh token
				rotatedRefreshToken, err := auth.RotateRefreshToken(cfg.DB, w, r, refreshTokenCookie)
				if err != nil { // in other words: if we have a refresh token cookie but the corresponding token is not working
					// if refresh token is broken, best to redirect to login
					auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
					http.Redirect(w, r, "/api/login", http.StatusSeeOther)
					return
				}

				// 1.4 make a new JWT (access token) based on refresh token
				newAccessToken, err := auth.MakeJWT(rotatedRefreshToken.UserID, "user", cfg.Secret, cfg.AccessTokenDuration)
				if err != nil {
					// this should probably not redirect anywhere and just cancel the whole ordeal - this is pretty fundamental
					jsonutils.WriteError(w, http.StatusInternalServerError, err, "error making new JWT")
					return
				}

				// 1.5 Set new cookies
				auth.SetAuthCookies(w, newAccessToken, rotatedRefreshToken.Token, "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)

				// 1.6 retry same request (redirect to original path. note that this time we will have proper cookies so 2.2 should trigger)
				http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
				return
			} else { // so if accessErr != nil && refreshErr != nil, meaning no cookie was found for either access or refresh token
				ctx := r.Context()
				ctx = context.WithValue(ctx, auth.UserIDContextKey, "") // this is to prevent an attacker sending a request without the cookie but with the UserIDContextKey manually set
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		}

		// 2. if accessErr == nil ... so we can deal with a functioning access token
		// first check if we have a refresh token
		refreshTokenCookie, refreshErr := r.Cookie("user_refresh_token")
		if refreshErr != nil {
			auth.SetAuthCookies(w, "", "", usertype, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
			http.Redirect(w, r, "/api/login", http.StatusSeeOther)
			return
		}

		// 2.1 then validate access token
		userID, ut, err := auth.ValidateJWT(accessTokenCookie.Value, cfg.GetSecret())
		// 2.1.1 sanity check
		if ut != usertype {
			// unexpected state: access token usertype does not match access token cookie name > clear cookies and sent to login
			jsonutils.WriteError(w, http.StatusBadRequest, errors.New("access token usertype does not match expectation from cookie name"), "access token usertype does not match cookie name")
			http.Redirect(w, r, "/api/login", http.StatusSeeOther)
			return
		}
		// 2.1.2 if the token is invalid auth.ValidateJWT will return an error
		if err != nil {
			// 2.1.3 get the refresh token
			rotatedRefreshToken, err := auth.RotateRefreshToken(cfg.DB, w, r, refreshTokenCookie)
			if err != nil { // in other words: if we have a refresh token cookie but the corresponding token is not working
				// if refresh token is broken, best to redirect to login
				auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
				http.Redirect(w, r, "/api/login", http.StatusSeeOther)
				return
			}

			// 2.1.4 make a new JWT (access token) based on refresh token
			newAccessToken, err := auth.MakeJWT(rotatedRefreshToken.UserID, "user", cfg.Secret, cfg.AccessTokenDuration)
			if err != nil {
				// this should probably not redirect anywhere and just cancel the whole ordeal - this is pretty fundamental
				jsonutils.WriteError(w, http.StatusInternalServerError, err, "error making new JWT")
				return
			}

			// 2.1.5 Set new cookies
			auth.SetAuthCookies(w, newAccessToken, rotatedRefreshToken.Token, "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)

			// 2.1.6 retry same request (redirect to original path. note that this time we will have proper cookies so 2.2 should trigger)
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return

		}

		// 2.2 rotate refresh token
		rotatedRefreshToken, err := auth.RotateRefreshToken(cfg.DB, w, r, refreshTokenCookie)
		if err != nil {
			if err == auth.ErrRefreshTokenInvalid {
				// unexpected state: valid access token but invalid refresh token > reset both cookies and redirect to /api/login
				auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
				http.Redirect(w, r, "/api/login", http.StatusSeeOther)
				return
			} else { // bit redundant but for clarity
				return // in this case we had an error querying the database. Don't want to redirect because it indicates a much bigger problem.
			}

		}

		// 2.3 set rotated refresh token to cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    rotatedRefreshToken.Token,
			Path:     "/api/refresh",
			Expires:  time.Now().Add(time.Duration(cfg.RefreshTokenDuration) * 24 * time.Hour), // one week for refresh tokens?
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		// 2.4 check if user is active
		var uid string
		user, err := cfg.DB.GetUserByID(r.Context(), userID)
		if errors.Is(err, sql.ErrNoRows) { // unexpected state: a non-existing user is specified in the access token jwt
			auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
			http.Redirect(w, r, "/api/login", http.StatusSeeOther)
			return
		} else if err != nil {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetUserByID in AuthUserMiddleware)")
			return
		}
		if !user.IsActive { // NYI
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}

		// 3. modify context to take ID and pass into next handler
		uid = userID.String()
		ctx := r.Context()
		ctx = context.WithValue(ctx, auth.UserIDContextKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (cfg *ApiConfig) AuthVisitorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/*
			Middleware for visitor authentication. Significantly simpler than user authentication as visitors do not use refresh tokens (for now).
			Returns a handler. As such, no error values are returned: instead we either redirect or write to log.
		*/
		// 0. init
		const usertype string = "visitor"

		// 1. retrieve visitor access cookie
		accessTokenCookie, err := r.Cookie("visitor_access_token")
		if err != nil {
			// err means no cookie found under this name. So serve as though not a visitor.
			next.ServeHTTP(w, r)
			return
		}

		// 2. validate visitor JWT
		visitorID, ut, err := auth.ValidateJWT(accessTokenCookie.Value, cfg.Secret)
		if ut != usertype { // sanity check
			// unexpected state: delete visitor cookie and pass
			auth.SetAuthCookies(w, "", "", "visitor", 2*cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
		}
		if err != nil { // I strongly doubt whether this is correct
			if err == jwt.ErrTokenExpired {
				// ? do I want something here?
			}
			next.ServeHTTP(w, r)
			return
		}

		// 3. add to context and pass to next
		ctx := r.Context()
		ctx = context.WithValue(ctx, auth.VisitorIDContextKey, visitorID.String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

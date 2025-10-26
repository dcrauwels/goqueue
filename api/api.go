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
	"github.com/google/uuid"
)

type ApiConfig struct {
	DB                   *database.Queries
	Secret               string
	Env                  string
	AccessTokenDuration  int
	RefreshTokenDuration int
	PublicIDGenerator    func() string
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
				rotatedRefreshToken, err := auth.RotateRefreshToken(cfg.DB, w, r, refreshTokenCookie) // note that auth.RotateRefreshToken() does a lot of heavy lifting here and checks token validity
				if err != nil {
					// this error is thrown if there is an issue with the refresh token provided in the corresponding cookie. auth.RotateRefreshToken() already calls jsonutils.WriteError()
					auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
					return
				}

				// 1.4 make a new JWT (access token) based on refresh token
				newAccessToken, err := auth.MakeJWT(rotatedRefreshToken.UserID, "user", cfg.Secret, cfg.AccessTokenDuration)
				if err != nil {
					// if this fails, there is a problem with issueing access tokens in general, which is very fundamental
					auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
					jsonutils.WriteError(w, http.StatusInternalServerError, err, "error making new JWT (MakeJWT in AuthUserMiddleware)")
					return
				}

				// 1.5 Set new cookies
				auth.SetAuthCookies(w, newAccessToken, rotatedRefreshToken.Token, "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)

				// 1.6 pass on to next handler
				uid := rotatedRefreshToken.UserID.String()
				ctx := r.Context()
				ctx = context.WithValue(ctx, auth.UserIDContextKey, uid)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else { // so if accessErr != nil && refreshErr != nil, meaning no cookie was found for either access or refresh token
				// 1.7 pass on to next handler with explicitly empty authentication
				ctx := r.Context()
				ctx = context.WithValue(ctx, auth.UserIDContextKey, "") // this is to prevent an attacker sending a request without the cookie but with the UserIDContextKey manually set
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		}

		// 2. if accessErr == nil ... so we can deal with a functioning access token
		// first check if we have a refresh token
		refreshTokenCookie, refreshErr := r.Cookie("user_refresh_token")
		if refreshErr != nil { // no refresh token cookie found
			auth.SetAuthCookies(w, "", "", usertype, cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
			jsonutils.WriteError(w, http.StatusUnauthorized, refreshErr, "access token but no refresh token found. Nulling auth cookies.")
			return
		}

		// 2.1 then validate access token
		userID, ut, err := auth.ValidateJWT(accessTokenCookie.Value, cfg.GetSecret())
		// 2.1.1 sanity check
		if ut != usertype {
			// unexpected state: access token usertype does not match access token cookie name > clear cookies and sent to login
			jsonutils.WriteError(w, http.StatusBadRequest, errors.New("access token usertype does not match expectation from cookie name"), "access token usertype does not match cookie name")
			return
		}
		// 2.1.2 if the access token is invalid auth.ValidateJWT will return an error
		if err != nil {
			// 2.1.3 get the refresh token
			rotatedRefreshToken, err := auth.RotateRefreshToken(cfg.DB, w, r, refreshTokenCookie)
			if err != nil { // in other words: if we have a refresh token cookie but the corresponding token is not working
				// if refresh token is broken, best to redirect to login
				auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
				return
			}

			// 2.1.4 no error at 2.1.3 means we have a valid refresh token > make a new access token based on refresh token
			newAccessToken, err := auth.MakeJWT(rotatedRefreshToken.UserID, "user", cfg.Secret, cfg.AccessTokenDuration)
			if err != nil {
				// this should probably not redirect anywhere and just cancel the whole ordeal - this is pretty fundamental
				jsonutils.WriteError(w, http.StatusInternalServerError, err, "error making new JWT")
				return
			}

			// 2.1.5 Set new cookies
			auth.SetAuthCookies(w, newAccessToken, rotatedRefreshToken.Token, "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)

			// 2.1.6 retry same request (redirect to original path. note that this time we will have proper cookies so 2.2 should trigger)
			uid := rotatedRefreshToken.UserID.String()
			ctx := r.Context()
			ctx = context.WithValue(ctx, auth.UserIDContextKey, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
			return

		}

		// 2.2 rotate refresh token
		rotatedRefreshToken, err := auth.RotateRefreshToken(cfg.DB, w, r, refreshTokenCookie) // note that auth.RotateRefreshToken calls WriteError on its own
		if err != nil {
			if err == auth.ErrRefreshTokenInvalid {
				// unexpected state: valid access token but invalid refresh token > reset both cookies and redirect to /api/login
				auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
				return
			} else { // bit redundant but for clarity
				// should we panic here? probably not
				return // in this case we had an error querying the database. Don't want to redirect because it indicates a much bigger problem.
			}

		}

		// 2.3 set rotated refresh token to cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    rotatedRefreshToken.Token,
			Path:     "/api/refresh",
			Expires:  time.Now().Add(time.Duration(cfg.RefreshTokenDuration) * 24 * time.Hour), // note that cfg.RefreshTokenDuration denotes duration in days
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		// 2.4 check if user is active
		var uid string
		user, err := cfg.DB.GetUserByID(r.Context(), userID)
		if errors.Is(err, sql.ErrNoRows) { // unexpected state: a non-existing user is specified in the access token jwt
			auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
			jsonutils.WriteError(w, http.StatusUnauthorized, err, "non-existing user specified in access token JWT. Note that this error should never occur.")
			return
		} else if err != nil {
			auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetUserByID in AuthUserMiddleware)")
			return
		}
		if !user.IsActive { // NYI
			auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
			jsonutils.WriteJSON(w, http.StatusForbidden, "user account described in authentication cookies is not active")
			return
		}

		// 3. modify context to take ID and pass into next handler
		uid = userID.String()
		ctx := r.Context()
		ctx = context.WithValue(ctx, auth.UserIDContextKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

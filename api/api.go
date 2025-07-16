package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/google/uuid"
)

type ApiConfig struct {
	DB                  *database.Queries
	Secret              string
	Env                 string
	AccessTokenDuration int
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

func (cfg *ApiConfig) makeAuthMiddleware(ck auth.ContextKey) func(next http.Handler) http.Handler {
	/*
		Abstracted function for making AuthUserMiddleware and AuthVisitorMiddleware.
		Returns a middleware function (i.e. a function with the following signature: func(next http.Handler) http.Handler).
		As this returns a funciton that returns a handler, no errors are returned. Instead, they are printed to stdout and
		sent to the writer as a response with corresponding HTTP status code.
	*/
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. check for access token cookie
			usertype := strings.TrimSuffix(ck.String(), "ID") // the ContextKeys are all in a "userID", "visitorID" format. So this just returns "user" etc. Note this might turn out to be restrictive at some point.
			accessTokenCookie, accessErr := r.Cookie(fmt.Sprintf("%s_access_token", usertype))
			if accessErr != nil { // implicitly this means the error returned is http.ErrNoCookie, as request.Cookie can only return that specific error
				// 1.1 if no access token cookie, check for refresh token cookie
				refreshTokenCookie, refreshErr := r.Cookie(fmt.Sprintf("%s_refresh_token", usertype))
				// 1.2 if refresh token cookie present: reissue access token cookie
				if refreshErr == nil { // if accessErr != nil && refreshErr == nil (both errors are guaranteed to be http.ErrNoCookie so this just means
					// there is no cookie with the "access_token" name but there is a "refresh_token" cookie)
					// 1.3 rotate refresh token
					rotatedRefreshToken, err := auth.RotateRefreshToken(cfg.DB, w, r, refreshTokenCookie)
					if err != nil {
						return
					}

					// 1.4 make a new JWT
					newAccessToken, err := auth.MakeJWT(rotatedRefreshToken.UserID, "user", cfg.Secret, cfg.AccessTokenDuration)

					// 1.5 Set new cookies
					http.SetCookie(w, &http.Cookie{
						Name:     "access_token",
						Value:    newAccessToken,
						Path:     "/",
						Expires:  time.Now().Add(time.Duration(cfg.AccessTokenDuration) * time.Minute),
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteLaxMode,
					})
					http.SetCookie(w, &http.Cookie{
						Name:     "refresh_token",
						Value:    rotatedRefreshToken.Token,
						Path:     "/api/refresh",
						Expires:  time.Now().Add(7 * 24 * time.Hour), // one week for refresh tokens?
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteStrictMode,
					})
					// 1.6 retry same request (redirect to original path. note that this time we will have proper cookies so ??2.?? will trigger)
					http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
					return
				}

			}
			// 2. if accessErr == nil ... so we can deal with a functioning access token
			authID, _, err := auth.ValidateJWT(accessTokenCookie.Value, cfg.GetSecret())
			// 2.1 if the token is invalid auth.ValidateJWT will return an error
			if err != nil {
				jsonutils.WriteError(w, http.StatusUnauthorized, err, "invalid or expired access token in cookie") // currently this is quite vague, not sure whether that's better for security purposes or I should split out different error types in auth.ValidateJWT()
				http.Redirect(w, r, "/login", http.StatusUnauthorized)
				return
			}

			// 2.2 rotate refresh token
			refreshTokenCookie, err := r.Cookie("user_refresh_token")
			if err != nil {
				http.Redirect(w, r, "/api/login", http.StatusSeeOther) // is this actually meaningful? or shoudl i just return and raise http.StatusUnauthorized?
				return
			}
			rotatedRefreshToken, err := auth.RotateRefreshToken(cfg.DB, w, r, refreshTokenCookie)
			if err != nil {
				return
			}

			// 2.3 set rotated refresh token to cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "refresh_token",
				Value:    rotatedRefreshToken.Token,
				Path:     "/api/refresh",
				Expires:  time.Now().Add(7 * 24 * time.Hour), // one week for refresh tokens?
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			})

			// 3. modify context to take ID and pass into next handler
			ctx := r.Context()
			ctx = context.WithValue(ctx, ck, authID.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (cfg *ApiConfig) AuthUserMiddleware(next http.Handler) http.Handler {
	userMiddleware := cfg.makeAuthMiddleware(auth.UserIDContextKey)
	return userMiddleware(next)
}

func (cfg *ApiConfig) AuthVisitorMiddleware(next http.Handler) http.Handler {
	visitorMiddleware := cfg.makeAuthMiddleware(auth.VisitorIDContextKey)
	return visitorMiddleware(next)
}

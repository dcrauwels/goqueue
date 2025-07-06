package api

import (
	"context"
	"database/sql"
	"net/http"

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

func (cfg *ApiConfig) AuthMiddleware(next http.Handler) http.Handler {
	// middleware for handling authentication through an access token and refresh token cookie
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. check for access token cookie
		accessTokenCookie, accessErr := r.Cookie("access_token")
		if accessErr != nil {
			// 1.1 if no access token cookie, check for refresh token cookie
			if accessErr == http.ErrNoCookie { // what about the situation where this is expired??
				refreshTokenCookie, refreshErr := r.Cookie("refresh_token")
				// 1.2 if refresh token cookie present: reissue access token cookie
				if refreshErr == nil {
					// 1.3 get the userID from the old refresh token to rotate the refresh token and make a new JWT
					oldRefreshToken, err := cfg.DB.RevokeRefreshTokenByToken(r.Context(), refreshTokenCookie.Value) // note that this already revokes the token ... is that bad?
					if err == sql.ErrNoRows {
						jsonutils.WriteError(w, 404, err, "no rows found in database matching refresh token")
						return
					} else if err != nil {
						jsonutils.WriteError(w, 500, err, "error querying database (GetRefreshTokenByToken in AuthMiddleware)")
						return
					}
					// 1.3 make a new JWT
					newAccessToken, err := auth.MakeJWT(oldRefreshToken.UserID, "user", cfg.Secret, cfg.AccessTokenDuration)

					// 1.4 rotate refresh token
					newRefreshToken, err := auth.MakeRefreshToken()
					if err != nil {
						jsonutils.WriteError(w, 500, err, "error creating refresh token (in AuthMiddleware)")
						return
					}
					rtParams := database.CreateRefreshTokenParams{
						Token:  newRefreshToken,
						UserID: oldRefreshToken.UserID,
					}
					rotatedRefreshToken, err := cfg.DB.CreateRefreshToken(r.Context())

				}

			}
		}

	})
}

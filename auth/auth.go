package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/google/uuid"
)

var ErrWrongUserType = errors.New("usertype supplied in JWT is not valid")

var ErrVisitorMismatch = errors.New("accessing visitor is not visitor identified in endpoint URI")

var ErrUserInactive = errors.New("user account is inactive")

type configReader interface {
	GetSecret() string
}

type databaseQueryer interface {
	GetUserByID(context.Context, uuid.UUID) (database.User, error)
	GetVisitorByID(context.Context, uuid.UUID) (database.Visitor, error)
	CreateRefreshToken(context.Context, database.CreateRefreshTokenParams) (database.RefreshToken, error)
	RevokeRefreshTokenByToken(context.Context, string) (database.RefreshToken, error)
}

type ContextKey string

func (ck ContextKey) String() string {
	return string(ck)
}

const UserIDContextKey ContextKey = "userID"
const VisitorIDContextKey ContextKey = "visitorID"

func SetAuthCookies(w http.ResponseWriter, accessToken, refreshToken, expectedAuthType string, accessTokenMinuteDuration, refreshTokenDayDuration int) {
	// Access Token Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     fmt.Sprintf("%s_access_token", expectedAuthType),
		Value:    accessToken,
		Path:     "/",
		Expires:  time.Now().Add(time.Duration(accessTokenMinuteDuration) * time.Minute), // currently set to 1 hour but refer to api.handlerloginuser and api.handlerrefreshuser in handler_auth.go
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Refresh Token Cookie
	if expectedAuthType == "user" {
		http.SetCookie(w, &http.Cookie{
			Name:     fmt.Sprintf("%s_refresh_token", expectedAuthType),
			Value:    refreshToken,
			Path:     "/api/refresh",
			Expires:  time.Now().Add(time.Duration(refreshTokenDayDuration) * 24 * time.Hour),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
	}

}

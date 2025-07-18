package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
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

func VisitorsByID(w http.ResponseWriter, r *http.Request, cfg configReader, db databaseQueryer) (uuid.UUID, error) {
	// boilerplate for GET and PUT /api/visitors/{visitor_id}
	// 1. read visitor ID from endpoint URI
	pv := r.PathValue("visitor_id")
	visitorID, err := uuid.Parse(pv)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "endpoint is not a valid ID")
		return visitorID, err
	}

	// 2. retrieve authentication data from cookies (through XFromContext functions)
	// 2.1 get user from cookie
	accessingUser, err := UserFromContext(w, r, db)
	if err != nil {
		return visitorID, err
	}

	// 2.2 get visitor from cookie

	// 3. authenticate: either for visitor with matching ID or user (both from JWT in 2)
	switch {
	case "user": // auth for user
		accessingParty, err := db.GetUserByID(r.Context(), accessingID)
		if err == sql.ErrNoRows {
			jsonutils.WriteError(w, 404, err, "accessing user does not exist in database")
			return visitorID, err
		} else if err != nil {
			jsonutils.WriteError(w, 500, err, "error querying database (GetUserByID)")
			return visitorID, err
		} else if !accessingParty.IsActive {
			jsonutils.WriteError(w, 403, ErrWrongUserType, "accessing user account is inactive")
			return visitorID, ErrWrongUserType
		}
	case "visitor": // auth for visitor
		accessingParty, err := db.GetVisitorByID(r.Context(), accessingID)
		if err == sql.ErrNoRows {
			jsonutils.WriteError(w, 404, err, "accessing visitor does not exist in database")
			return visitorID, err
		} else if err != nil {
			jsonutils.WriteError(w, 500, err, "error querying database (GetVisitorByID)")
			return visitorID, err
		} else if accessingParty.ID != visitorID { // so the visitor is trying to edit a different visitor. not allowed
			jsonutils.WriteError(w, 403, ErrVisitorMismatch, "accessing visitor is not requested visitor")
			return visitorID, ErrVisitorMismatch
		} else if accessingParty.ID != accessingID { // sanity check
			jsonutils.WriteError(w, 500, ErrVisitorMismatch, "accessing visitor is not corresponding to database")
			return visitorID, ErrVisitorMismatch
		}
	default: // in case usertype is neither "user" nor "visitor"
		jsonutils.WriteError(w, 400, ErrWrongUserType, "incorrect usertype in JWT")
		return visitorID, ErrWrongUserType
	}

	return visitorID, nil
}

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

package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/google/uuid"
)

type authDependencies interface {
	GetUserByID(context.Context, uuid.UUID) (database.User, error)
	GetSecret() string
}

func UserFromHeader(w http.ResponseWriter, r *http.Request, deps authDependencies) (database.User, error) {
	// use this function when you are performing an action that assumes a logged in user
	// returns the user in question
	// return user from header authentication
	accessingUser := database.User{}
	accessToken, err := GetBearerToken(r.Header)
	if err != nil {
		jsonutils.WriteError(w, 401, err, "no authorization field in header")
		return accessingUser, err
	}
	//validate token
	accessingUserID, err := ValidateJWT(accessToken, deps.GetSecret())
	if err != nil {
		jsonutils.WriteError(w, 401, err, "access token invalid")
		return accessingUser, err
	}
	//query for user by ID and run checks
	accessingUser, err = deps.GetUserByID(r.Context(), accessingUserID)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "user not found")
		return accessingUser, err
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return accessingUser, err
	}

	return accessingUser, nil
}

func IsAdminFromHeader(w http.ResponseWriter, r *http.Request, deps authDependencies) (bool, error) {
	// return IsAdmin bool from header authentication
	accessingUser, err := UserFromHeader(w, r, deps)
	if err != nil {
		return false, err
	} else if !accessingUser.IsAdmin {
		err = errors.New("user not authorized")
		jsonutils.WriteError(w, 401, err, "missing IsAdmin status")
		return false, err
	}

	return true, nil
}

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

func AdminFromHeader(w http.ResponseWriter, r *http.Request, deps authDependencies) {

	// check for admin status in accessing user
	//get access token
	accessToken, err := GetBearerToken(r.Header)
	if err != nil {
		jsonutils.WriteError(w, 401, err, "no authorization field in header")
		return
	}
	//validate token
	accessingUserID, err := ValidateJWT(accessToken, deps.GetSecret())
	if err != nil {
		jsonutils.WriteError(w, 401, err, "access token invalid")
		return
	}
	//query for user by ID and run checks
	accessingUser, err := deps.GetUserByID(r.Context(), accessingUserID)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "user not found")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return
	} else if !accessingUser.IsAdmin {
		jsonutils.WriteError(w, 401, errors.New("user not authorized"), "missing IsAdmin status")
		return
	}
}

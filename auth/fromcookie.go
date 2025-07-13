package auth

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/google/uuid"
)

var ErrNoIDInContext = errors.New("auth: no ID provided in context")

func UserFromContext(w http.ResponseWriter, r *http.Request, db databaseQueryer) (database.User, error) {
	/*
		Parses the HTTP request context for a user. This is done by checking for a value at the UserIDContextKey key. Then queries database and returns the user.
		Possible errors:
		- auth.ErrNoIDInContext ("auth: no ID provided in context")
		- sql.ErrNoRows
		-
	*/

	// 1. init
	user := database.User{}

	// 2. get UserIDContextKey value from context
	userIDString, ok := r.Context().Value(UserIDContextKey).(string)
	if !ok {
		jsonutils.WriteError(w, http.StatusUnauthorized, ErrNoIDInContext, "no user ID provided in context (in auth.UserFromContext)")
		return user, ErrNoIDInContext
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "user ID in context is not a valid UUID (in auth.UserFromContext)")
		return user, err
	}

	// 3. query database for user
	user, err = db.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "user not found in database (in auth.UserFromContext)")
			return user, err
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetUserByID in auth.UserFromContext)")
			return user, err
		}
	}

	return user, nil
}

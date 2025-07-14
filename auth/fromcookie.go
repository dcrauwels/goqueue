package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/google/uuid"
)

var ErrNoIDInContext = errors.New("auth: no ID provided in context")

func authFromContext[T any](
	w http.ResponseWriter,
	r *http.Request,
	ck contextKey,
	expectedAuthType string,
	GetByID func(context.Context, uuid.UUID) (T, error),
) (T, error) {
	/*
		Parses the HTTP request context for a user or visitor. This is done by checking for a value at the contextKey.
		Then queries database and returns the user/visitor. Implemented in auth.UserFromContext and auth.VisitorFromContext.
		Possible errors:
		- auth.ErrNoIDInContext ("auth: no ID provided in context")
		- sql.ErrNoRows
	*/
	// 1. init
	var accessor T

	// 2. get contextKey value from context
	IDString, ok := r.Context().Value(ck).(string)
	if !ok {
		jsonutils.WriteError(w, http.StatusUnauthorized, ErrNoIDInContext, fmt.Sprintf("no %s ID provided in context (in auth.AuthFromContext)", expectedAuthType))
		return accessor, ErrNoIDInContext
	}
	// 2.1 and parse as UUID
	ID, err := uuid.Parse(IDString)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, fmt.Sprintf("%s ID in context is not a valid UUID (in auth.AuthFromContext)", expectedAuthType))
		return accessor, err
	}

	// 3. query DB for accessor
	accessor, err = GetByID(r.Context(), ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, fmt.Sprintf("%s not found in database (GetByID in auth.AuthFromContext)", expectedAuthType))
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetByID in auth.AuthFromContext)")
		}
		return accessor, err
	}

	// 4. return
	return accessor, nil
}

func UserFromContext(w http.ResponseWriter, r *http.Request, db databaseQueryer) (database.User, error) {
	/*
		Implements auth.authFromContext for user authentication from cookie.
	*/
	var user database.User
	user, err := authFromContext(w, r, UserIDContextKey, "user", db.GetUserByID)
	if err != nil {
		return user, err
	}
	return user, err
}

func VisitorFromContext(w http.ResponseWriter, r *http.Request, db databaseQueryer) (database.Visitor, error) {
	/*
		Implements auth.authFromContext for visitor authentication from cookie.
		NOTE THE BACKEND FOR THIS COOKIE TYPE IS NYI
	*/
	var visitor database.Visitor
	visitor, err := authFromContext(w, r, VisitorIDContextKey, "visitor", db.GetVisitorByID)
	if err != nil {
		return visitor, err
	}
	return visitor, err
}

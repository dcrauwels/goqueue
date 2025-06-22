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

type configReader interface {
	GetSecret() string
}

type databaseQueryer interface {
	GetUserByID(context.Context, uuid.UUID) (database.User, error)
	GetVisitorByID(context.Context, uuid.UUID) (database.Visitor, error)
}

var ErrWrongUserType = errors.New("wrong usertype")

func UserTypeFromHeader(w http.ResponseWriter, r *http.Request, cfg configReader) (string, error) {
	// short function to retrieve purely the usertype from the JWT in the header authorization field
	accessToken, err := GetBearerToken(r.Header)
	if err != nil {
		jsonutils.WriteError(w, 401, err, "no authorization field in request header")
		return "", err
	}
	_, userType, err := ValidateJWT(accessToken, cfg.GetSecret())
	if err != nil {
		jsonutils.WriteError(w, 403, err, "access token invalid")
		return "", err
	}
	return userType, nil
}

func AuthFromHeader[T any](
	w http.ResponseWriter,
	r *http.Request,
	cfg configReader,
	expectedUserType string,
	queryFunc func(context.Context, uuid.UUID) (T, error),
) (T, error) {
	// generic function to generate UserFromHeader and VisitorFromHeader without having to write the same function twice
	var zero T
	accessToken, err := GetBearerToken(r.Header)
	if err != nil {
		jsonutils.WriteError(w, 401, err, "no authorization field in request header")
		return zero, err
	}

	//validate token
	id, userType, err := ValidateJWT(accessToken, cfg.GetSecret())
	if err != nil {
		jsonutils.WriteError(w, 403, err, "access token invalid")
		return zero, err
	} else if userType != "user" {
		jsonutils.WriteError(w, 400, err, "not logged in as "+expectedUserType)
		return zero, ErrWrongUserType
	}

	//query for user by ID and run checks
	entity, err := queryFunc(r.Context(), id)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, expectedUserType+" not found")
		return zero, err
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return zero, err
	}

	return entity, nil
}

func UserFromHeader(w http.ResponseWriter, r *http.Request, cfg configReader, db databaseQueryer) (database.User, error) {
	return AuthFromHeader(w, r, cfg, "user", db.GetUserByID)
}

func VisitorFromHeader(w http.ResponseWriter, r *http.Request, cfg configReader, db databaseQueryer) (database.Visitor, error) {
	return AuthFromHeader(w, r, cfg, "visitor", db.GetVisitorByID)
}

func IsAdminFromHeader(w http.ResponseWriter, r *http.Request, cfg configReader, db databaseQueryer) (bool, error) {
	// return IsAdmin bool from header authentication
	accessingUser, err := UserFromHeader(w, r, cfg, db)
	if err != nil {
		return false, err
	}
	return accessingUser.IsAdmin, nil
}

package auth

import (
	"context"
	"database/sql"
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

func UserFromHeader(w http.ResponseWriter, r *http.Request, cfg configReader, db databaseQueryer) (database.User, error) {
	// use this function when you are performing an action that assumes a logged in user
	// returns the user in question
	// return user from header authentication
	accessingUser := database.User{}
	accessToken, err := GetBearerToken(r.Header)
	if err != nil {
		jsonutils.WriteError(w, 401, err, "no authorization field in request header")
		return accessingUser, err
	}
	//validate token
	accessingUserID, userType, err := ValidateJWT(accessToken, cfg.GetSecret())
	if err != nil {
		jsonutils.WriteError(w, 403, err, "access token invalid")
		return accessingUser, err
	} else if userType != "user" {
		jsonutils.WriteError(w, 400, err, "not logged in as user")
		return accessingUser, err
	}
	//query for user by ID and run checks
	accessingUser, err = db.GetUserByID(r.Context(), accessingUserID)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "user not found")
		return accessingUser, err
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return accessingUser, err
	}

	return accessingUser, nil
}

func VisitorFromHeader(w http.ResponseWriter, r *http.Request, cfg configReader, db databaseQueryer) (database.Visitor, error) {
	// same idea as UserFromHeader(). Bit of repetition in first 10 lines or so
	// this may turn out to be a bad idea.
	accessingVisitor := database.Visitor{}
	accessToken, err := GetBearerToken(r.Header)
	if err != nil {
		jsonutils.WriteError(w, 403, err, "no authorization field in request header")
		return accessingVisitor, err
	}

	//validate token
	accessingVisitorID, userType, err := ValidateJWT(accessToken, cfg.GetSecret())
	if err != nil {
		jsonutils.WriteError(w, 400, err, "access token invalid")
		return accessingVisitor, err
	} else if userType != "visitor" {
		jsonutils.WriteError(w, 403, err, "not logged in as visitor")
		return accessingVisitor, err
	}

	//query for visitor by ID and run checks
	accessingVisitor, err = db.GetVisitorByID(r.Context(), accessingVisitorID)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "visitor not found")
		return accessingVisitor, err
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database (GetVisitorByID)")
		return accessingVisitor, err
	}

	return accessingVisitor, nil
}

func IsAdminFromHeader(w http.ResponseWriter, r *http.Request, cfg configReader, db databaseQueryer) (bool, error) {
	// return IsAdmin bool from header authentication
	accessingUser, err := UserFromHeader(w, r, cfg, db)
	if err != nil {
		return false, err
	}
	return accessingUser.IsAdmin, nil
}

package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/dcrauwels/goqueue/strutils"
	"github.com/google/uuid"
)

func (cfg *ApiConfig) HandlerPostUsers(w http.ResponseWriter, r *http.Request) { // POST /admin/users
	// check for admin status in accessing user
	//get access token
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		jsonutils.WriteError(w, 401, err, "no authorization field in header")
		return
	}
	//validate token
	accessingUserID, err := auth.ValidateJWT(accessToken, cfg.Secret)
	if err != nil {
		jsonutils.WriteError(w, 401, err, "access token invalid")
		return
	}
	//query for user by ID and run checks
	accessingUser, err := cfg.DB.GetUserByID(r.Context(), accessingUserID)
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

	// get request data
	decoder := json.NewDecoder(r.Body)
	reqParams := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	err = decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "JSON formatting invalid")
		return
	}

	// check request for validity
	//email valid
	if err = strutils.ValidateEmail(reqParams.Email); err != nil {
		jsonutils.WriteError(w, 400, err, "email formatting invalid: please use jdoe@provider.tld.")
		return
	}
	//password valid (aA0)
	if err = strutils.ValidatePassword(reqParams.Password); err != nil {
		jsonutils.WriteError(w, 400, err, "password formatting invalid: please use lowercase, uppercase and/or numeric, between 8 and 30 characters.")
		return
	}

	// hash password
	hashedPassword, err := auth.HashPassword(reqParams.Password)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "password could not be hashed.")
		return
	}

	// run query CreateUser
	queryParams := database.CreateUserParams{
		Email:          reqParams.Email,
		HashedPassword: hashedPassword,
	}
	createdUser, err := cfg.DB.CreateUser(r.Context(), queryParams)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "could not query database to create user.")
	}

	// return response 201
	responseParams := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		IsAdmin   bool      `json:"is_admin"`
		IsActive  bool      `json:"is_active"`
	}{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
		IsAdmin:   createdUser.IsAdmin,
		IsActive:  createdUser.IsActive,
	}
	jsonutils.WriteJSON(w, 201, responseParams)

}

func (cfg *ApiConfig) HandlerPutUsers(w http.ResponseWriter, r *http.Request) { // PUT /api/users
	// check for current user ID

}

func (cfg *ApiConfig) HandlerDeleteUsers(w http.ResponseWriter, r *http.Request) { // DELETE /api/users
	// check for admin status in current user
	// run query DeleteUserByID
	return
}

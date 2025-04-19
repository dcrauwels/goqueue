package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/strutils"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPostUsers(w http.ResponseWriter, r *http.Request) { // POST /admin/users
	// check for admin status in current user
	//NYI
	// get request data
	decoder := json.NewDecoder(r.Body)
	reqParams := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}
	err := decoder.Decode(&reqParams)
	if err != nil {
		writeError(w, 400, err, "JSON formatting invalid")
		return
	}

	// check request for validity
	//email valid
	if err = strutils.ValidateEmail(reqParams.Email); err != nil {
		writeError(w, 400, err, "email formatting invalid: please use jdoe@provider.tld.")
		return
	}
	//password valid (aA0)
	if err = strutils.ValidatePassword(reqParams.Password); err != nil {
		writeError(w, 400, err, "password formatting invalid: please use lowercase, uppercase and/or numeric, between 8 and 30 characters.")
		return
	}

	// hash password
	hashedPassword, err := auth.HashPassword(reqParams.Password)
	if err != nil {
		writeError(w, 500, err, "password could not be hashed.")
		return
	}

	// run query CreateUser
	queryParams := database.CreateUserParams{
		Email:          reqParams.Email,
		HashedPassword: hashedPassword,
	}
	user, err := cfg.db.CreateUser(r.Context(), queryParams)
	if err != nil {
		writeError(w, 500, err, "could not query database to create user.")
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
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		IsActive:  user.IsActive,
	}
	writeJSON(w, 201, responseParams)

}

func (cfg *apiConfig) handlerPutUsers(w http.ResponseWriter, r *http.Request) { // PUT /api/users
	// check for current user ID

}

func (cfg *apiConfig) handlerDeleteUsers(w http.ResponseWriter, r *http.Request) { // DELETE /api/users
	// check for admin status in current user
	// run query DeleteUserByID
	return
}

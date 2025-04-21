package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/dcrauwels/goqueue/strutils"
	"github.com/google/uuid"
)

type usersRequestParameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type usersResponseParameters struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"is_admin"`
	IsActive  bool      `json:"is_active"`
}

func processUsersParameters(w http.ResponseWriter, reqParams usersRequestParameters) (string, error) {
	// check request for validity
	//email valid
	if err := strutils.ValidateEmail(reqParams.Email); err != nil {
		jsonutils.WriteError(w, 400, err, "password formatting invalid: please use jdoe@provider.tld")
		return "", err
	}
	//password valid (aA0)
	if err := strutils.ValidatePassword(reqParams.Password); err != nil {
		jsonutils.WriteError(w, 400, err, "password formatting invalid: please use lowercase, uppercase and/or numeric, between 8 and 30 characters.")
		return "", err
	}

	// hash password
	hashedPassword, err := auth.HashPassword(reqParams.Password)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "password could not be hashed.")
		return "", err
	}

	return hashedPassword, nil
}

func (cfg *ApiConfig) HandlerPostUsers(w http.ResponseWriter, r *http.Request) { // POST /api/users
	// check for admin status in accessing user
	userIsAdmin, err := auth.IsAdminFromHeader(w, r, cfg)
	if err != nil || !userIsAdmin {
		// already used jsonutils.WriteError in the auth.IsAdminFromHeader function. No need to repeat here
		return
	}

	// get request data
	decoder := json.NewDecoder(r.Body)
	reqParams := usersRequestParameters{}
	err = decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "JSON formatting invalid")
		return
	}

	// check request for validity & hash password
	hashedPassword, err := processUsersParameters(w, reqParams)
	if err != nil {
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
	responseParams := usersResponseParameters{
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
	// 1. check for admin status in accessing user
	accessingUser, err := auth.UserFromHeader(w, r, cfg)
	if err != nil {
		// already used jsonutils.WriteError in the auth.UserFromHeader function. No need to repeat here
		return
	}

	// 2. get request data
	decoder := json.NewDecoder(r.Body)
	reqParams := usersRequestParameters{}
	err = decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "JSON formatting invalid")
		return
	}

	// 3. check for validity and prep hashed password
	hashedPassword, err := processUsersParameters(w, reqParams)
	if err != nil {
		return
	}

	// 4. run query

	//

	// 5. write response

}

func (cfg *ApiConfig) HandlerDeleteUsers(w http.ResponseWriter, r *http.Request) { // DELETE /api/users
	// check for admin status in accessing user
	userIsAdmin, err := auth.IsAdminFromHeader(w, r, cfg)
	if err != nil || !userIsAdmin {
		// already used jsonutils.WriteError in the auth.IsAdminFromHeader function. No need to repeat here
		return
	}

	// run query DeleteUserByID
	return
}

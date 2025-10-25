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
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type UsersRequestParameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type UsersAdminRequestParameters struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	IsAdmin  bool   `json:"is_admin"`
	IsActive bool   `json:"is_active"`
}

type UsersResponseParameters struct {
	ID        uuid.UUID `json:"id"`
	PublicID  string    `json:"public_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	IsAdmin   bool      `json:"is_admin"`
	IsActive  bool      `json:"is_active"`
}

func (urp *UsersResponseParameters) Populate(u database.User) {
	urp.ID = u.ID
	urp.PublicID = u.PublicID
	urp.CreatedAt = u.CreatedAt
	urp.UpdatedAt = u.UpdatedAt
	urp.Email = u.Email
	urp.FullName = u.FullName
	urp.IsAdmin = u.IsAdmin
	urp.IsActive = u.IsActive
}

func ProcessUsersParameters(w http.ResponseWriter, request UsersRequestParameters) (string, error) {
	/*
		This function checks if the parameters in request (email, password and full name) are valid for use in an INSERT query to the users table.
		Returns a hashed password (using auth.HashPassword) and an error. If the function fails, an empty string is returned instead.
	*/

	//email valid
	if err := strutils.ValidateEmail(request.Email); err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "password formatting invalid: please use jdoe@provider.tld")
		return "", err
	}
	//password valid (aA0)
	if err := strutils.ValidatePassword(request.Password); err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "password formatting invalid: please use lowercase, uppercase and/or numeric, between 8 and 30 characters.")
		return "", err
	}

	// hash password
	hashedPassword, err := auth.HashPassword(request.Password)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "password could not be hashed.")
		return "", err
	}

	return hashedPassword, nil
}

// POST /api/users (admin only)
func (cfg *ApiConfig) HandlerPostUsers(w http.ResponseWriter, r *http.Request) {
	// function to CREATE new user
	// 1. check for admin status in accessing user
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		return
	} else if !accessingUser.IsAdmin {
		jsonutils.WriteError(w, http.StatusUnauthorized, ErrNotAdmin, "non-admin user tried to request POST /api/users")
		return
	}

	// 2. get request data
	decoder := json.NewDecoder(r.Body)
	reqParams := UsersRequestParameters{}
	err = decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "JSON formatting invalid")
		return
	}

	// 3. check request for validity & hash password
	hashedPassword, err := ProcessUsersParameters(w, reqParams) // this function already calls jsonutils.WriteError, no need to do so here
	if err != nil {
		return
	}

	// 4. generate publicid
	pid, err := gonanoid.New(cfg.PublicIDLength)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error generating nanoid (gonanoid.New in HandlerPostUsers)")
		return
	}

	// 5. run query CreateUser
	queryParams := database.CreateUserParams{
		PublicID:       pid,
		Email:          reqParams.Email,
		HashedPassword: hashedPassword,
		FullName:       reqParams.FullName,
	}
	createdUser, err := cfg.DB.CreateUser(r.Context(), queryParams)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "could not query database to create user.")
		return
	}

	// return response 201
	response := UsersResponseParameters{}
	response.Populate(createdUser)

	jsonutils.WriteJSON(w, http.StatusCreated, response)

}

// PUT /api/users
func (cfg *ApiConfig) HandlerPutUsers(w http.ResponseWriter, r *http.Request) { // PUT /api/users
	/*
		Function to change own details for user. Only things a user can change about himself are email, password and fullname
		Currently the way this is set up is that a user changes himself. But perhaps it would be better to only keep the PUT /api/users/{user_id} setup and
		remove this endpoint.
	*/

	// 1. get accessing user from context
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access PUT /api/users")
		return
	}

	// 2. get request data
	decoder := json.NewDecoder(r.Body)
	reqParams := UsersRequestParameters{}
	err = decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "JSON formatting invalid")
		return
	}

	// 3. check for validity and prep hashed password
	hashedPassword, err := ProcessUsersParameters(w, reqParams)
	if err != nil {
		return
	}

	// 4. run query
	queryParams := database.SetUserEmailPasswordByIDParams{
		ID:             accessingUser.ID,
		Email:          reqParams.Email,
		HashedPassword: hashedPassword,
	}
	updatedUser, err := cfg.DB.SetUserEmailPasswordByID(r.Context(), queryParams)
	if errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusForbidden, err, "user does not exist. How did you do this?")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database")
		return
	}

	// 5. write response
	response := UsersResponseParameters{}
	response.Populate(updatedUser)
	jsonutils.WriteJSON(w, http.StatusOK, response)

}

// PUT /api/users/{public_user_id}
func (cfg *ApiConfig) HandlerPutUsersByID(w http.ResponseWriter, r *http.Request) {
	// function to UPDATE specific user by public ID
	// requires isadmin status from accessing user

	// 1. retrieve accessing user
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access PUT /api/users")
		return
	}

	// 2. retrieve target user from uri
	pid := r.PathValue("public_user_id")
	if len(pid) != cfg.PublicIDLength {
		jsonutils.WriteError(w, http.StatusBadRequest, errors.New("incorrect public ID length"), "invalid public ID length provided in endpoint")
		return
	}

	// 3. retrieve request data
	request := UsersAdminRequestParameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "invalid json request structure")
		return
	}

	// 4. auth: either IsAdmin or userID match
	if !accessingUser.IsAdmin {
		if accessingUser.PublicID != pid {
			jsonutils.WriteError(w, http.StatusForbidden, errors.New("user not authorized to send a PUT request to this endpoint"), "this user account is not authorized to send a PUT request to this endpoint")
			return
		} else if (request.IsAdmin && !accessingUser.IsAdmin) || (request.IsActive != accessingUser.IsActive) { // non-admins cannot set themselves to admin obviously or (de)activate themselves
			jsonutils.WriteError(w, http.StatusForbidden, errors.New("user not authorized to edit these fields on own account"), "this user account cannot edit their own admin or activity status")
			return
		}
	}

	// 5. run query
	queryParams := database.SetUserByPublicIDParams{
		PublicID: pid,
		Email:    request.Email,
		FullName: request.FullName,
		IsAdmin:  request.IsAdmin,
		IsActive: request.IsActive,
	}
	updatedUser, err := cfg.DB.SetUserByPublicID(r.Context(), queryParams)
	if errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusNotFound, err, "user not found")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database")
		return
	}

	// 5. write response
	response := UsersResponseParameters{}
	response.Populate(updatedUser)
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

func (cfg *ApiConfig) HandlerGetUsers(w http.ResponseWriter, r *http.Request) { // GET /api/users
	// READs all users
	// requires isadmin status from accessing user

	// 1. check if accessing user is admin
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access GET /api/users")
		return
	} else if !accessingUser.IsAdmin {
		jsonutils.WriteError(w, http.StatusForbidden, err, "GET /api/users is only accessible to admin level users")
		return
	}

	// 2. run query
	users, err := cfg.DB.GetUsers(r.Context())
	if errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusNotFound, err, "no users found")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database")
		return
	}

	// 3. write response
	response := make([]UsersResponseParameters, len(users))
	for i, u := range users {
		response[i].Populate(u)
	}
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

func (cfg *ApiConfig) HandlerGetUsersByID(w http.ResponseWriter, r *http.Request) { // GET /api/users/{public_user_id}
	/*
		Handler function to retrieve a full user (including is_admin) based on the UUID. This needs to be accessible to all clients, even unauthenticated
		ones, because a visitor needs to be able to see who is calling him. I might decide to change this at a later date though.
	*/

	// 1. check authentication from context
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if errors.Is(err, auth.ErrNoIDInContext) {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access this endpoint")
		return
	} else if err != nil {
		return // any other err than auth.ErrNoIDInContext already sends a json.WriteError so no additional error writing is needed
	}
	// 1.1 sanity checks
	if !accessingUser.IsActive {
		jsonutils.WriteError(w, http.StatusForbidden, err, "accessing user account is inactive")
		return
	}

	// 2. get user ID from request uri
	pid := r.PathValue("public_user_id")

	// 3. check for validity
	if len(pid) != cfg.PublicIDLength {
		jsonutils.WriteError(w, http.StatusBadRequest, errors.New("incorrect public ID length"), "invalid public ID length provided in endpoint")
		return
	}

	// 4. run query
	user, err := cfg.DB.GetUserByPublicID(r.Context(), pid)
	if errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusNotFound, err, "user not found")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database(GetUserByPublicID in HandlerGetUsersByID)")
		return
	}

	// 5. write response
	response := UsersResponseParameters{}
	response.Populate(user)
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

// not entirely sure how I want to go about this function yet
/*func (cfg *ApiConfig) HandlerDeleteUsers(w http.ResponseWriter, r *http.Request) { // DELETE /api/users
	// 1. get accessing user
	accessingUser, err := auth.UserFromHeader(w, r, cfg)
	if err != nil {
		return
	}

	// 2. check if userID was provided, otherwise delete self
	targetUserID := accessingUser.ID

	// 2. get user from response
	reqParams := UsersRequestParameters{Password: ""}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "incorrect json request provided")
	}

	// 3. run query DeleteUserByID

	deletedUser, err := cfg.DB.DeleteUserByID(r.Context(), reqParams.Email)

	// 3. write response

}*/

package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/dcrauwels/goqueue/strutils"
	"github.com/google/uuid"
)

type UsersRequestParameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type UsersResponseParameters struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	IsAdmin   bool      `json:"is_admin"`
	IsActive  bool      `json:"is_active"`
}

func (urp *UsersResponseParameters) Populate(u database.User) {
	urp.ID = u.ID
	urp.CreatedAt = u.CreatedAt
	urp.UpdatedAt = u.UpdatedAt
	urp.Email = u.Email
	urp.FullName = u.FullName
	urp.IsAdmin = u.IsAdmin
	urp.IsActive = u.IsActive
}

func ProcessUsersParameters(w http.ResponseWriter, reqParams UsersRequestParameters) (string, error) {
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
	// function to CREATE new user
	// check for admin status in accessing user
	userIsAdmin, err := auth.IsAdminFromHeader(w, r, cfg)
	if err != nil || !userIsAdmin {
		// already used jsonutils.WriteError in the auth.IsAdminFromHeader function. No need to repeat here
		return
	}

	// get request data
	decoder := json.NewDecoder(r.Body)
	reqParams := UsersRequestParameters{}
	err = decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "JSON formatting invalid")
		return
	}

	// check request for validity & hash password
	hashedPassword, err := ProcessUsersParameters(w, reqParams)
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
		return
	}

	// return response 201
	response := UsersResponseParameters{}
	response.Populate(createdUser)

	jsonutils.WriteJSON(w, 201, response)

}

func (cfg *ApiConfig) HandlerPutUsers(w http.ResponseWriter, r *http.Request) { // PUT /api/users
	// function to change own details for user. Only things a user can change about himself are email, password and fullname
	// NYI: fullname editing

	// 1. get accessing user from header
	accessingUser, err := auth.UserFromHeader(w, r, cfg)
	if err != nil {
		// already used jsonutils.WriteError in the auth.UserFromHeader function. No need to repeat here
		return
	}

	// 2. get request data
	decoder := json.NewDecoder(r.Body)
	reqParams := UsersRequestParameters{}
	err = decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "JSON formatting invalid")
		return
	}

	// 3. check for validity and prep hashed password
	hashedPassword, err := ProcessUsersParameters(w, reqParams)
	if err != nil {
		return
	}

	// 4. run query
	queryParams := database.SetEmailPasswordByIDParams{
		ID:             accessingUser.ID,
		Email:          reqParams.Email,
		HashedPassword: hashedPassword,
	}
	updatedUser, err := cfg.DB.SetEmailPasswordByID(r.Context(), queryParams)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 403, err, "user does not exist. How did you do this?")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return
	}

	// 5. write response
	response := UsersResponseParameters{}
	response.Populate(updatedUser)
	jsonutils.WriteJSON(w, 200, response)

}

func (cfg *ApiConfig) HandlerGetUsers(w http.ResponseWriter, r *http.Request) { // GET /api/users
	// returns list of all users
	// requires isadmin status from accessing user

	// 1. check if accessing user is admin
	isAdmin, err := auth.IsAdminFromHeader(w, r, cfg)
	if err != nil {
		// auth.UserFromHeader() already calls jsonutils.WriteError()
		return
	} else if !isAdmin {
		jsonutils.WriteError(w, 403, err, "GET /api/users is only accessible to admin level users")
		return
	}
	// 2. run query
	users, err := cfg.DB.GetUsers(r.Context())
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "no users found")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return
	}
	// 3. write response
	response := make([]UsersResponseParameters, len(users))
	for i, u := range users {
		response[i] = UsersResponseParameters{}
		response[i].Populate(u)
	}
	jsonutils.WriteJSON(w, 200, response)
}

func (cfg *ApiConfig) HandlerGetUsersByID(w http.ResponseWriter, r *http.Request) { // GET /api/users/{user_id}
	// 1. get user ID from request uri
	req := r.PathValue("user_id")

	// 2. check for validity
	userID, err := uuid.Parse(req)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "endpoint is not a valid user ID")
		return
	}

	// 3. run query
	user, err := cfg.GetUserByID(r.Context(), userID)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "user not found")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return
	}
	// 4. write response
	response := UsersResponseParameters{}
	response.Populate(user)
	jsonutils.WriteJSON(w, 200, response)
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
		jsonutils.WriteError(w, 400, err, "incorrect json request provided")
	}

	// 3. run query DeleteUserByID

	deletedUser, err := cfg.DB.DeleteUserByID(r.Context(), reqParams.Email)

	// 3. write response

}*/

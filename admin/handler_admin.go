package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dcrauwels/goqueue/api"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/google/uuid"
)

type configReader interface {
	GetSecret() string
	GetEnv() string
	GeneratePublicID() string
}

type databaseQueryer interface {
	CreateUser(context.Context, database.CreateUserParams) (database.User, error)
	SetUserIsAdminByID(context.Context, database.SetUserIsAdminByIDParams) (database.User, error)
	GetUserByID(context.Context, uuid.UUID) (database.User, error)
}

func AdminCreateUser(w http.ResponseWriter, r *http.Request, cfg configReader, db databaseQueryer) {
	// used for making an admin auth user
	// 1. check dev env (there is no point checking isAdmin)
	env := cfg.GetEnv()
	if env != "dev" {
		jsonutils.WriteError(w, 403, errors.New("endpoint not approached in DEV enrironment"), "incorrect environment")
		return
	}

	// 2. get req params: email & password
	request := api.UsersPOSTRequestParameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "incorrect json request structure")
		return
	}
	hashedPassword, err := api.ProcessUsersParameters(w, request)
	if err != nil {
		return //already calls jsonutils.WriteError
	}

	// 3. make user
	pid := cfg.GeneratePublicID()
	queryCreateParams := database.CreateUserParams{
		Email:          request.Email,
		HashedPassword: hashedPassword,
		FullName:       request.FullName,
		PublicID:       pid,
	}
	createdUser, err := db.CreateUser(r.Context(), queryCreateParams)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database for user creation")
		return
	}

	// 4. set user admin
	queryAdminParams := database.SetUserIsAdminByIDParams{
		ID:      createdUser.ID,
		IsAdmin: true,
	}
	adminUser, err := db.SetUserIsAdminByID(r.Context(), queryAdminParams)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database for setting IsAdmin")
		return
	}

	// 5. response
	response := api.UsersResponseParameters{
		ID:        adminUser.ID,
		CreatedAt: adminUser.CreatedAt,
		UpdatedAt: adminUser.UpdatedAt,
		Email:     adminUser.Email,
		IsAdmin:   adminUser.IsAdmin,
		IsActive:  adminUser.IsActive,
	}
	jsonutils.WriteJSON(w, 200, response)
}

//func AdminDeleteUser(w http.ResponseWriter, r *http.Request, cfg adminCfgDependencies) {
// used for fully deleting a user from the database. admin env only.
// 1. check dev environment
// 2. read request for user ID
// 3. run query for deletion
// 4. respond 204
//}

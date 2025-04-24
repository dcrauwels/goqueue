package admin

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dcrauwels/goqueue/api"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
)

type adminCfgDependencies interface {
	CreateUser(context.Context, database.CreateUserParams) (database.User, error)
	SetIsAdminByID(context.Context, database.SetIsAdminByIDParams) (database.User, error)
}

func AdminCreateUser(w http.ResponseWriter, r *http.Request, cfg adminCfgDependencies) {
	// used for making an admin auth user
	// 1. check admin status
	// 1a. check dev env

	// 2. get req params: email & password
	reqParams := api.UsersRequestParameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "incorrect json request structure")
		return
	}
	hashedPassword, err := api.ProcessUsersParameters(w, reqParams)
	if err != nil {
		return
	}

	// 3. make user
	queryCreateParams := database.CreateUserParams{
		Email:          reqParams.Email,
		HashedPassword: hashedPassword,
	}
	createdUser, err := cfg.CreateUser(r.Context(), queryCreateParams)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database for user creation")
		return
	}

	// 4. set user admin
	queryAdminParams := database.SetIsAdminByIDParams{
		ID:      createdUser.ID,
		IsAdmin: true,
	}
	adminUser, err := cfg.SetIsAdminByID(r.Context(), queryAdminParams)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database for setting IsAdmin")
		return
	}

	// 5. response
	respParams := api.UsersResponseParameters{
		ID:        adminUser.ID,
		CreatedAt: adminUser.CreatedAt,
		UpdatedAt: adminUser.UpdatedAt,
		Email:     adminUser.Email,
		IsAdmin:   adminUser.IsAdmin,
		IsActive:  adminUser.IsActive,
	}
	jsonutils.WriteJSON(w, 200, respParams)
}

func AdminDeleteUser(w http.ResponseWriter, r *http.Request, cfg adminCfgDependencies) {
	// used for fully deleting a user from the database. admin env only.
	// 1. check admin status

	// 2. read request for user ID
	// 3. run query for deletion
	// 4. respond 204
}

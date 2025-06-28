package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
)

func purposeChanger[T struct{}](
	w http.ResponseWriter,
	r *http.Request,
	cfg *ApiConfig,
	queryFunc func(context.Context, T) database.Purpose) {
	// 1. auth for access: user, isadmin
	isAdmin, err := auth.IsAdminFromHeader(w, r, cfg, cfg.DB)
	if err != nil {
		return
	} else if !isAdmin {
		jsonutils.WriteError(w, 403, err, "non-admin tried to access POST or PUT /api/purposes")
		return
	}
	// 2. read request
	decoder := json.NewDecoder(r.Body)
	request := T{}
	err = decoder.Decode(&request)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "user provided incorrect JSON data to POST or PUT /api/purposes")
		return
	}

	// 3. sanity check? probably if parent purpose ID is in DB
	if request.ParentPurposeID.valid {
		_, err := cfg.DB.GetPurposesByID(r.Context())
	}

	// 4. query database
	// 5. write response

}

func (cfg *ApiConfig) HandlerPostPurposes(w http.ResponseWriter, r *http.Request) {
	// 1. auth for access: user, isadmin
	isAdmin, err := auth.IsAdminFromHeader(w, r, cfg, cfg.DB)
	if err != nil {
		return
	} else if !isAdmin {
		jsonutils.WriteError(w, 403, err, "non-admin tried to access POST or PUT /api/purposes")
		return
	}

	// 2. read request
	decoder := json.NewDecoder(r.Body)
	request := database.CreatePurposeParams{}
	err = decoder.Decode(&request)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "user provided invalid JSON in a PUT request to /api/purposes")
		return
	}

	// 3. query database
	createdPurpose, err := cfg.DB.CreatePurpose(r.Context(), request)
	// 4. write response

}

// POST and PUT can be unified into a single function

func (cfg *ApiConfig) HandlerPutPurposesByID(w http.ResponseWriter, r *http.Request) {
	// 1. auth for access: user, isadmin
	isAdmin, err := auth.IsAdminFromHeader(w, r, cfg, cfg.DB)
	if err != nil {
		return
	} else if !isAdmin {
		jsonutils.WriteError(w, 403, err, "non-admin tried to access POST or PUT /api/purposes")
		return
	}

	// 2. read request

	// 3. sanity check: if parent purpose ID is in DB
	// 4. query database
	// 5. write response

}

func (cfg *ApiConfig) HandlerGetPurposes(w http.ResponseWriter, r *http.Request) {

}

func (cfg *ApiConfig) HandlerGetPurposesByID(w http.ResponseWriter, r *http.Request) {

}

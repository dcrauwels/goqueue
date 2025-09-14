package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/google/uuid"
)

type PurposesRequestParameters struct {
	PurposeName     string        `json:"purpose_name"`
	ParentPurposeID uuid.NullUUID `json:"parent_purpose_id"`
}

type PurposesResponseParameters struct {
	ID              uuid.UUID     `json:"id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	PurposeName     string        `json:"purpose_name"`
	ParentPurposeID uuid.NullUUID `json:"parent_purpose_id"`
}

func (prp *PurposesResponseParameters) Populate(p database.Purpose) {
	prp.ID = p.ID
	prp.CreatedAt = p.CreatedAt
	prp.UpdatedAt = p.UpdatedAt
	prp.PurposeName = p.PurposeName
	prp.ParentPurposeID = p.ParentPurposeID
}

var ErrNotAdmin = errors.New("user does not have admin status")

// Helper function that handles the common logic
func (cfg *ApiConfig) handlePurposeOperation(
	w http.ResponseWriter,
	r *http.Request,
	operation string, // http operation name (POST, PUT, GET etc.) for error messages
	requestPtr any, // pointer to request parameter struct (like PurposesPutRequestParameters etc.)
	dbQuery func() (database.Purpose, error), // function to execute the database query, so either cfg.DB.CreatePurpose() or cfg.DB.SetPurpose()
) {
	/*
		This function provides boilerplate for both PUT and POST operations to the /api/purposes endpoint.
	*/
	// 1. auth for access: user, isadmin
	user, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, fmt.Sprintf("user authorization is required to request %s /api/purposes", operation))
		return
	} else if !user.IsAdmin {
		jsonutils.WriteError(w, http.StatusForbidden, ErrNotAdmin, fmt.Sprintf("non-admin user tried to request %s /api/purposes", operation))
		return
	}

	// 2. read request (delegated to caller)
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&requestPtr)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, fmt.Sprintf("user provided invalid JSON in a request to %s /api/purposes", operation))
		return
	}

	// 3. query database (delegated to caller)
	result, err := dbQuery()
	if errors.Is(err, sql.ErrNoRows) {
		switch operation {
		case "POST":
			jsonutils.WriteError(w, http.StatusBadRequest, err, "user provided invalid parent_purpose_id when requesting POST /api/purposes")
			return
		case "PUT":
			jsonutils.WriteError(w, http.StatusBadRequest, err, "user provided invalid id or parent_purpose_id when requesting PUT /api/purposes")
			return
		}
	} else if err != nil {
		var dbFuncName string
		switch operation {
		case "POST":
			dbFuncName = "CreatePurpose"
		case "PUT":
			dbFuncName = "SetPurpose"
		}

		jsonutils.WriteError(w, http.StatusInternalServerError, err, fmt.Sprintf("error querying database (%s)", dbFuncName))
		return
	}

	// 4. write response
	response := PurposesResponseParameters{}
	response.Populate(result)
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

// POST /api/purposes (admin only)
func (cfg *ApiConfig) HandlerPostPurposes(w http.ResponseWriter, r *http.Request) {
	var request PurposesRequestParameters // note that we only need to inituate a PurposeRequestParameters struct and it is populated by handlePurposeOperation

	cfg.handlePurposeOperation(w, r, "POST",
		&request,
		// Database operation function
		func() (database.Purpose, error) {
			queryParams := database.CreatePurposeParams{
				PurposeName:     request.PurposeName,
				ParentPurposeID: request.ParentPurposeID,
			}
			return cfg.DB.CreatePurpose(r.Context(), queryParams)
		},
	)
}

// PUT /api/purposes/{purpose_id} (admin only)
func (cfg *ApiConfig) HandlerPutPurposesByID(w http.ResponseWriter, r *http.Request) {
	var request PurposesRequestParameters // note that we only need to inituate a PurposeRequestParameters struct and it is populated by handlePurposeOperation

	// retrieve request ID
	req := r.PathValue("user_id")
	purposeID, err := uuid.Parse(req)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "endpoint is not a valid user ID")
		return
	}

	cfg.handlePurposeOperation(w, r, "PUT",
		// Decoder function
		func(interface{}) error {
			decoder := json.NewDecoder(r.Body)
			return decoder.Decode(&request)
		},
		// Database operation function
		func() (database.Purpose, error) {
			queryParams := database.SetPurposeParams{
				ID:              purposeID,
				PurposeName:     request.PurposeName,
				ParentPurposeID: request.ParentPurposeID,
			}
			return cfg.DB.SetPurpose(r.Context(), queryParams)
		},
	)
}

// GET /api/purposes (no authentication or request body required)
func (cfg *ApiConfig) HandlerGetPurposes(w http.ResponseWriter, r *http.Request) {
	// 1. run query
	purposes, err := cfg.DB.GetPurposes(r.Context())
	if errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusNotFound, err, "no purposes found in database when requesting GET /api/purposes")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetPurposes)")
		return
	}

	// 2. write response
	response := make([]PurposesResponseParameters, len(purposes))
	for i, u := range purposes {
		response[i].Populate(u)
	}
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

func (cfg *ApiConfig) HandlerGetPurposesByID(w http.ResponseWriter, r *http.Request) {
	// 1. get purpose ID from endpoint
	req := r.PathValue("purpose_id")
	purposeID, err := uuid.Parse(req)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "endpoint is not a valid UUID")
		return
	}

	// 2. run query
	purpose, err := cfg.DB.GetPurposesByID(r.Context(), purposeID)
	if errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusNotFound, err, "no purposes found in database when requesting GET /api/purpose/{purpose_id}")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetPurposesByID)")
		return
	}

	// 3. write response
	var response PurposesResponseParameters
	response.Populate(purpose)
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

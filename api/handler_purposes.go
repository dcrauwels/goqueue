package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/google/uuid"
)

type PurposesPostRequestParameters struct {
	PurposeName     string        `json:"purpose_name"`
	ParentPurposeID uuid.NullUUID `json:"parent_purpose_id"`
}

type PurposesPutRequestParameters struct {
	ID uuid.UUID `json:"id"`
	PurposesPostRequestParameters
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

// Helper function that handles the common logic
func (cfg *ApiConfig) handlePurposeOperation(
	w http.ResponseWriter,
	r *http.Request,
	operation string, // http operation name (POST, PUT, GET etc.) for error messages
	requestPtr any, // pointer to request parameter struct (like PurposesPutRequestParameters etc.)
	dbQuery func() (database.Purpose, error), // function to execute the database query
) {
	// 1. auth for access: user, isadmin
	isAdmin, err := auth.IsAdminFromHeader(w, r, cfg, cfg.DB)
	if err != nil {
		return
	} else if !isAdmin {
		jsonutils.WriteError(w, 403, err, fmt.Sprintf("non-admin user tried to request %s /api/purposes", operation))
		return
	}

	// 2. read request (delegated to caller)
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&requestPtr)
	if err != nil {
		jsonutils.WriteError(w, 400, err, fmt.Sprintf("user provided invalid JSON in a request to %s /api/purposes", operation))
		return
	}

	// 3. query database (delegated to caller)
	result, err := dbQuery()
	if err == sql.ErrNoRows {
		switch operation {
		case "POST":
			jsonutils.WriteError(w, 400, err, "user provided invalid parent_purpose_id when requesting POST /api/purposes")
			return
		case "PUT":
			jsonutils.WriteError(w, 400, err, "user provided invalid id or parent_purpose_id when requesting PUT /api/purposes")
			return
		}
	} else if err != nil {
		dbFuncName := ""
		switch operation {
		case "POST":
			dbFuncName = "CreatePurpose"
		case "PUT":
			dbFuncName = "SetPurpose"
		}

		jsonutils.WriteError(w, 500, err, fmt.Sprintf("error querying database (%s)", dbFuncName))
		return
	}

	// 4. write response
	response := PurposesResponseParameters{}
	response.Populate(result)
	jsonutils.WriteJSON(w, 200, response)
}

// POST /api/purposes (admin only)
func (cfg *ApiConfig) HandlerPostPurposes(w http.ResponseWriter, r *http.Request) {
	var request PurposesPostRequestParameters

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

// POST /api/purposes (admin only)
func (cfg *ApiConfig) HandlerPutPurposesByID(w http.ResponseWriter, r *http.Request) {
	var request PurposesPutRequestParameters

	cfg.handlePurposeOperation(w, r, "PUT",
		// Decoder function
		func(interface{}) error {
			decoder := json.NewDecoder(r.Body)
			return decoder.Decode(&request)
		},
		// Database operation function
		func() (database.Purpose, error) {
			queryParams := database.SetPurposeParams{
				ID:              request.ID,
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
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "no purposes found when requesting GET /api/purposes")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database (GetPurposes)")
		return
	}

	// 2. write response
	response := make([]PurposesResponseParameters, len(purposes))
	for i, u := range response {
		u.Populate(purposes[i])
	}
	jsonutils.WriteJSON(w, 200, response)
}

func (cfg *ApiConfig) HandlerGetPurposesByID(w http.ResponseWriter, r *http.Request) {
	// NYI do I even want this?
}

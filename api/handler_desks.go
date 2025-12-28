package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/dcrauwels/goqueue/strutils"
	"github.com/google/uuid"
)

type DesksPostRequestParameters struct {
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
}

type DesksPutRequestParameters struct {
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	IsActive    bool           `json:"is_active"`
}

type DesksResponseParameters struct {
	ID          uuid.UUID      `json:"id"`
	Description sql.NullString `json:"description"`
	IsActive    bool           `json:"is_active"`
	PublicID    string         `json:"public_id"`
	Name        string         `json:"name"`
}

func (drp *DesksResponseParameters) Populate(d database.Desk) {
	drp.ID = d.ID
	drp.Description = d.Description
	drp.IsActive = d.IsActive
	drp.PublicID = d.PublicID
	drp.Name = d.Name
}

// POST /api/desks
func (cfg *ApiConfig) HandlerPostDesks(w http.ResponseWriter, r *http.Request) {
	// 1. check auth -> admin only
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required for this endpoint")
		return
	} else if !accessingUser.IsActive {
		jsonutils.WriteError(w, http.StatusForbidden, auth.ErrUserInactive, "user not active")
		return

	} else if !accessingUser.IsAdmin {
		jsonutils.WriteError(w, http.StatusForbidden, auth.ErrUserNotAdmin, "user requires admin status for this endpoint")
		return
	}

	// 2. get request data
	req := DesksPostRequestParameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "JSON formatting invalid")
		return
	}

	// 3. generate public ID
	dpid := cfg.PublicIDGenerator()

	// 4. run query CreateDesks
	queryParams := database.CreateDesksParams{
		PublicID:    dpid,
		Name:        req.Name,
		Description: req.Description,
	}

	result, err := cfg.DB.CreateDesks(r.Context(), queryParams)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (CreateDesks in HandlerPostDesks)")
		return
	}

	// 5. return result
	response := DesksResponseParameters{}
	response.Populate(result)
	jsonutils.WriteJSON(w, http.StatusCreated, response)
}

// PUT /api/desks/{desk_public_id}
func (cfg *ApiConfig) HandlerPutDesksByPublicID(w http.ResponseWriter, r *http.Request) {
	// 1. check auth -> admin only
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required for this endpoint")
		return
	} else if !accessingUser.IsAdmin {
		jsonutils.WriteError(w, http.StatusForbidden, auth.ErrUserNotAdmin, "user requires admin status for this endpoint")
		return
	}

	// 2. get path value
	dpid, err := strutils.GetPublicIDFromPathValue("desk_public_id", cfg.PublicIDLength, r)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "incorrect path value length")
		return
	}

	// 3. get request body
	request := DesksPutRequestParameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "JSON formatting invalid")
		return
	}

	// 4. run query SetDesksByPublicID
	queryParams := database.SetDesksByPublicIDParams{
		PublicID:    dpid,
		Name:        request.Name,
		Description: request.Description,
		IsActive:    request.IsActive,
	}
	desk, err := cfg.DB.SetDesksByPublicID(r.Context(), queryParams)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no desks found at specified public id")
			return
		}
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (SetDesksByPublicID in HandlerPutDesksByPublicID)")
		return
	}

	// 5. return result
	response := DesksResponseParameters{}
	response.Populate(desk)
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

// GET /api/desks
func (cfg *ApiConfig) HandlerGetDesks(w http.ResponseWriter, r *http.Request) {
	// 1. check auth
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required for this endpoint")
		return
	} else if !accessingUser.IsActive {
		jsonutils.WriteError(w, http.StatusForbidden, err, "accessing user not active")
		return
	}

	// 2. run query GetDesks
	desks, err := cfg.DB.GetDesks(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no rows found (GetDesks in HandlerGetDesks)")
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetDesks in HandlerGetDesks)")
		}
		return
	}

	// 3. return result
	response := make([]DesksResponseParameters, len(desks))
	for i, d := range desks {
		response[i].Populate(d)
	}
	jsonutils.WriteJSON(w, http.StatusOK, response)

}

// GET /api/desks/{desk_public_id}
func (cfg *ApiConfig) HandlerGetDesksByPublicID(w http.ResponseWriter, r *http.Request) {
	// 1. get path value
	dpid, err := strutils.GetPublicIDFromPathValue("desk_public_id", cfg.PublicIDLength, r)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "incorrect path value length")
		return
	}

	// 2. run query GetDesksByPublicID
	desk, err := cfg.DB.GetDesksByPublicID(r.Context(), dpid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no rows found (GetDesksByPublicID in HandlerGetDesksByPublicID)")
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetDesksByPublicID in HandlerGetDesksByPublicID)")
		}
		return
	}

	// 3. return result
	response := DesksResponseParameters{}
	response.Populate(desk)
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

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
)

type ServicelogsPOSTRequestParameters struct {
	VisitorPublicID string `json:"visitor_public_id"`
	UserPublicID    string `json:"user_public_id"`
	DeskPublicID    string `json:"desk_public_id"`
}

type ServicelogsPUTRequestParameters struct {
	ServicelogsPOSTRequestParameters
	IsActive bool `json:"is_active"`
}

type ServicelogsResponseParameters struct {
	ID              uuid.UUID `json:"id"`
	PublicID        string    `json:"public_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	VisitorPublicID string    `json:"visitor_public_id"`
	UserPublicID    string    `json:"user_public_id"`
	DeskPublicID    string    `json:"desk_public_id"`
	CalledAt        time.Time `json:"called_at"`
	IsActive        bool      `json:"is_active"`
}

func (slrp *ServicelogsResponseParameters) Populate(sl database.ServiceLog) {
	slrp.ID = sl.ID
	slrp.PublicID = sl.PublicID
	slrp.CreatedAt = sl.CreatedAt
	slrp.UpdatedAt = sl.UpdatedAt
	slrp.VisitorPublicID = sl.VisitorPublicID
	slrp.UserPublicID = sl.UserPublicID
	slrp.DeskPublicID = sl.DeskPublicID
	slrp.CalledAt = sl.CalledAt
	slrp.IsActive = sl.IsActive
}

func handleServiceLogOperation[T any](
	w http.ResponseWriter,
	r *http.Request,
	operation string,
	requestPtr *T,
	dbQuery func() (database.ServiceLog, error),
) {

	// 1. read request
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(requestPtr)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "invalid JSON request")
		return
	}

	// 2. query DB
	serviceLog, err := dbQuery()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no service logs found at the provided public ID")
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (handleServiceLogOperation)")
		}
		return
	}

	// 3. write response
	response := ServicelogsResponseParameters{}
	response.Populate(serviceLog)
	var statusCode int
	if operation == "POST" {
		statusCode = http.StatusCreated
	} else {
		statusCode = http.StatusOK
	}
	jsonutils.WriteJSON(w, statusCode, response)

}

// POST /api/servicelogs (user only)
func (cfg *ApiConfig) HandlerPostServicelogs(w http.ResponseWriter, r *http.Request) {
	// 1. auth
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication is required for this endpoint")
		return
	} else if !accessingUser.IsActive {
		jsonutils.WriteError(w, http.StatusForbidden, err, "accessing user account is inactive")
		return
	}

	// 2. handleServiceLogOperation
	request := &ServicelogsPOSTRequestParameters{}
	handleServiceLogOperation(
		w,
		r,
		"POST",
		request,
		func() (database.ServiceLog, error) {
			query := database.CreateServiceLogsParams{
				PublicID:        cfg.PublicIDGenerator(),
				UserPublicID:    request.UserPublicID,
				VisitorPublicID: request.VisitorPublicID,
				DeskPublicID:    request.DeskPublicID,
			}
			return cfg.DB.CreateServiceLogs(r.Context(), query)
		},
	)
}

// PUT /api/servicelogs/{servicelog_id} (user only)
func (cfg *ApiConfig) HandlerPutServicelogsByID(w http.ResponseWriter, r *http.Request) {
	// 1. auth
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication is required for this endpoint")
		return
	} else if !accessingUser.IsActive {
		jsonutils.WriteError(w, http.StatusForbidden, err, "accessing user account is inactive")
		return
	}

	// 2. handleServiceLogOperation
	request := &ServicelogsPUTRequestParameters{}
	slpid, err := strutils.GetPublicIDFromPathValue("servicelog_public_id", cfg.PublicIDLength, r)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "invalid service log path")
		return
	}
	handleServiceLogOperation(
		w,
		r,
		"PUT",
		request,
		func() (database.ServiceLog, error) {
			query := database.SetServiceLogsByPublicIDParams{
				PublicID:        slpid,
				VisitorPublicID: request.VisitorPublicID,
				UserPublicID:    request.UserPublicID,
				DeskPublicID:    request.DeskPublicID,
				IsActive:        request.IsActive,
			}
			return cfg.DB.SetServiceLogsByPublicID(r.Context(), query)
		},
	)
}

// GET /api/servicelogs (user only)
func (cfg *ApiConfig) HandlerGetServicelogs(w http.ResponseWriter, r *http.Request) {
}

// GET /api/servicelogs/{servicelog_public_id}
func (cfg *ApiConfig) HandlerGetServicelogsByPublicID(w http.ResponseWriter, r *http.Request) {
	// 1. get path publicid
	slpid, err := strutils.GetPublicIDFromPathValue("servicelog_public_id", cfg.PublicIDLength, r)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "invalid service log path")
		return
	}

	// 2. run query
	servicelog, err := cfg.DB.GetServiceLogsByPublicID(r.Context(), slpid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no rows found at provided public id")
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetServiceLogsByPublicID in HandlerGetServiceLogsByPublicID)")
		}
		return
	}

	// 3. write response
	response := ServicelogsResponseParameters{}
	response.Populate(servicelog)
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

// GET /api/servicelogs/{visitor_id} not convinced this is needed
//func (cfg *ApiConfig) HandlerGetServicelogsByVisitorID(w http.ResponseWriter, r *http.Request) {
//}

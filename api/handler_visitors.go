package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/dcrauwels/goqueue/strutils"
	"github.com/google/uuid"
)

type VisitorsPostRequestParameters struct {
	Name            string `json:"name"`
	PurposePublicID string `json:"purpose_public_id"`
}

type VisitorsPutRequestParameters struct {
	PublicID        string `json:"public_id"`
	Name            string `json:"name"`
	PurposePublicID string `json:"purpose_public_id"`
	Status          int32  `json:"status"`
}

type VisitorsResponseParameters struct {
	ID                uuid.UUID      `json:"id"`
	PublicID          string         `json:"public_id"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	WaitingSince      time.Time      `json:"waiting_since"`
	Name              sql.NullString `json:"name"`
	PurposePublicID   string         `json:"purpose_public_id"`
	Status            int32          `json:"status"`
	DailyTicketNumber int32          `json:"daily_ticket_number"`
}

func (vrp *VisitorsResponseParameters) Populate(v database.Visitor) {
	vrp.ID = v.ID
	vrp.PublicID = v.PublicID
	vrp.CreatedAt = v.CreatedAt
	vrp.UpdatedAt = v.UpdatedAt
	vrp.WaitingSince = v.WaitingSince
	vrp.Name = v.Name
	vrp.PurposePublicID = v.PurposePublicID
	vrp.Status = v.Status
	vrp.DailyTicketNumber = v.DailyTicketNumber
}

// POST /api/visitors no auth required
func (cfg *ApiConfig) HandlerPostVisitors(w http.ResponseWriter, r *http.Request) { // POST /api/visitors
	/* function for sending a POST request to CREATE a single visitor from scratch
	in context the visitor accesses a website, enters his name and purpose and gets a number*/

	// 1. get request data: name, purpose
	decoder := json.NewDecoder(r.Body)
	request := VisitorsPostRequestParameters{}
	err := decoder.Decode(&request)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "JSON formatting invalid")
		return
	}

	// 2. check purpose for validity
	purpose, err := cfg.DB.GetPurposesByPublicID(r.Context(), request.PurposePublicID)
	if errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusNotFound, err, "purpose not found in database, please register first")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetPurposesByPublicID in HandlerPostVisitors)")
		return
	}

	// 3. query DB: UpdateTicketCounter
	dtn, err := cfg.DB.UpdateTicketCounter(r.Context())
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (UpdateTicketCounter in HandlerPostVisitors)")
		return
	}

	// 4. create public ID
	pid := cfg.PublicIDGenerator()

	// 4. query DB: CreateVisitor
	queryParams := database.CreateVisitorParams{
		PublicID:          pid,
		Name:              strutils.InitNullString(request.Name), // name is currently nullable.
		PurposePublicID:   purpose.PublicID,
		DailyTicketNumber: dtn,
	}

	createdVisitor, err := cfg.DB.CreateVisitor(r.Context(), queryParams)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (CreateVisitor in HandlerPostVisitors)")
		return
	}

	// 5. return response 201
	response := VisitorsResponseParameters{}
	response.Populate(createdVisitor)
	jsonutils.WriteJSON(w, http.StatusCreated, response)
}

func (cfg *ApiConfig) HandlerPutVisitorsByPublicID(w http.ResponseWriter, r *http.Request) { // PUT /api/visitors/{visitor_public_id}
	/*
		Handler function for dealing with PUT requests to the /api/visitors/{visitor_public_id} endpoint.
		Can be accessed only by users. While one can imagine cases where visitors want to edit their name
		after the fact (e.g. because of typos) I think the added value of allowing them to do so is minimal.
	*/

	// 1. get target visitor from URI
	pvid, err := strutils.GetPublicIDFromPathValue("visitor_public_id", cfg.PublicIDLength, r)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "incorrect path value length")
		return
	}

	// 2. get user authentication from context
	_, err = auth.UserFromContext(w, r, cfg.DB) // I don't need information about the user itself, just whether a user ID is present in the request context.
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access PUT /api/visitors")
		return
	}

	// 3. PUT request
	decoder := json.NewDecoder(r.Body)
	request := VisitorsPutRequestParameters{}
	err = decoder.Decode(&request)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "JSON formatting invalid")
		return
	}
	request.PublicID = pvid

	// 4. run query
	queryParams := database.SetVisitorByPublicIDParams{
		PublicID:        request.PublicID,
		Name:            strutils.InitNullString(request.Name),
		PurposePublicID: request.PurposePublicID,
		Status:          request.Status,
	}
	updatedVisitor, err := cfg.DB.SetVisitorByPublicID(r.Context(), queryParams)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "updated visitor does not exist in database")
			return
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (SetVisitorByID)")
			return
		}
	}

	// 6. write response
	response := VisitorsResponseParameters{}
	response.Populate(updatedVisitor)

	jsonutils.WriteJSON(w, http.StatusOK, response)
}

func (cfg *ApiConfig) HandlerGetVisitors(w http.ResponseWriter, r *http.Request) { // GET /api/visitors
	// only accessible to logged in users
	// 1. get user authentication from request context
	_, err := auth.UserFromContext(w, r, cfg.DB) // not interested in actual information about the user
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access GET /api/visitors")
		return
	}
	var visitors []database.Visitor

	// 2. check for query parameters (purpose, status)
	queryParameters := r.URL.Query()
	queryPurpose := queryParameters.Get("purpose")
	queryStatus := queryParameters.Get("status")

	// 2.1 status as string to status as int32
	status64, err := strconv.ParseInt(queryStatus, 10, 32)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "query parameter 'status' only takes integer values")
		return
	}
	status := int32(status64) // cast as int32 as the SQL query parameter structs take this
	// 2.2 check status for validity
	// NYI

	// 3. query database
	switch {
	case queryPurpose != "" && queryStatus != "": // both query parameters entered
		queryParams := database.GetVisitorsByPurposePublicIDAndStatusParams{
			PurposePublicID: queryPurpose,
			Status:          status,
		}
		visitors, err = cfg.DB.GetVisitorsByPurposePublicIDAndStatus(r.Context(), queryParams)
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no visitors found in database for purpose "+queryPurpose+" and status "+queryStatus)
			return
		} else if err != nil {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetVisitorsByPurposeStatus)")
			return
		}

	case queryPurpose != "" && queryStatus == "": // only purpose query parameter entered
		visitors, err = cfg.DB.GetVisitorsByPurposePublicID(r.Context(), queryPurpose)
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no visitors found in database for purpose "+queryPurpose)
			return
		} else if err != nil {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetVisitorsByPurpose)")
			return
		}

	case queryPurpose == "" && queryStatus != "": // only status query parameter entered
		visitors, err = cfg.DB.GetVisitorsByStatus(r.Context(), status)
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no viistors found in database for status "+queryStatus)
			return
		} else if err != nil {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetVisitorsByStatus)")
			return
		}

	case queryPurpose == "" && queryStatus == "": // neither query parameter entered
		visitors, err = cfg.DB.GetVisitors(r.Context())
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no visitors found in database")
			return
		} else if err != nil {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database")
			return
		}
	}

	// 4. write response
	response := make([]VisitorsResponseParameters, len(visitors))
	for i, u := range visitors {
		response[i].Populate(u)
	}
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

func (cfg *ApiConfig) HandlerGetVisitorsByPublicID(w http.ResponseWriter, r *http.Request) { // GET /api/visitors/{visitor_public_id}
	// 1. get visitor ID from endpoint
	pvid, err := strutils.GetPublicIDFromPathValue("visitor_public_id", cfg.PublicIDLength, r)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "incorrect path value length")
		return
	}

	// 2. get auth from context
	_, userErr := auth.UserFromContext(w, r, cfg.DB)                      // not interested in the actual user itself
	accessingVisitor, visitorErr := auth.VisitorFromContext(w, r, cfg.DB) // I think this is deprecated actually, see todo.md

	// 3. authenticate: either for visitor with matching ID or user
	if userErr != nil { // if userErr == nil then user authentication was provided and we are good to go
		if visitorErr != nil { // so userErr != nil && visitorErr != nil > no authentication whatsoever is provided
			jsonutils.WriteError(w, http.StatusUnauthorized, userErr, "authorization is required to access GET /api/visitors")
			return
		} else if pvid != accessingVisitor.PublicID { // visitors can only GET themselves
			jsonutils.WriteError(w, http.StatusForbidden, userErr, "visitors are only allowed to GET their own ID at /api/visitors")
			return
		}
	}

	// 3.1 should I put redundancy here for the user auth?
	// because now we just assume if userErr == nil everything is fine & dandy but that's a bit of a risk
	// NYI

	// 4. run query
	visitor, err := cfg.DB.GetVisitorsByPublicID(r.Context(), pvid)
	if errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusNotFound, err, "visitor not found in database")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetVisitorByID)")
		return
	}

	// 5. write response
	response := VisitorsResponseParameters{}
	response.Populate(visitor)
	jsonutils.WriteJSON(w, http.StatusOK, response)

}

package api

import (
	"database/sql"
	"encoding/json"
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
	Name      string    `json:"name"`
	PurposeID uuid.UUID `json:"purpose_id"`
}

type VisitorsPutRequestParameters struct {
	Name      string    `json:"name"`
	PurposeID uuid.UUID `json:"purpose_id"`
	Status    int32     `json:"status"`
}

type VisitorsResponseParameters struct {
	ID           uuid.UUID      `json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	WaitingSince time.Time      `json:"waiting_since"`
	Name         sql.NullString `json:"name"`
	PurposeID    uuid.UUID      `json:"purpose_id"`
	Status       int32          `json:"status"`
}

type VisitorsPOSTResponseParameters struct {
	VisitorsResponseParameters
	VisitorAccessToken string `json:"visitor_access_token"`
}

func (vrp *VisitorsResponseParameters) Populate(v database.Visitor) {
	vrp.ID = v.ID
	vrp.CreatedAt = v.CreatedAt
	vrp.UpdatedAt = v.UpdatedAt
	vrp.WaitingSince = v.WaitingSince
	vrp.Name = v.Name
	vrp.PurposeID = v.PurposeID
	vrp.Status = v.Status
}

func (cfg *ApiConfig) HandlerPostVisitors(w http.ResponseWriter, r *http.Request) { // POST /api/visitors
	// function for sending a POST request to CREATE a single visitor from scratch
	// in context the visitor accesses a website, enters his name and purpose and gets a number
	//
	// 1. get request data: name, purpose
	decoder := json.NewDecoder(r.Body)
	request := VisitorsPostRequestParameters{}
	err := decoder.Decode(&request)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "JSON formatting invalid")
		return
	}

	// 2. check purpose for validity
	purpose, err := cfg.DB.GetPurposesByID(r.Context(), request.PurposeID)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, http.StatusNotFound, err, "purpose not found in database, please register first")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetPurposesByName)")
		return
	}

	// 3. query DB: CreateVisitor
	queryParams := database.CreateVisitorParams{
		Name:      strutils.InitNullString(request.Name), // name is currently nullable.
		PurposeID: purpose.ID,
	}
	createdVisitor, err := cfg.DB.CreateVisitor(r.Context(), queryParams)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "could not query database to create visitor")
		return
	}

	// 4. make visitor access token
	visitorAccessToken, err := auth.MakeJWT(createdVisitor.ID, "visitor", cfg.Secret, 120)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "could not create access token")
		return
	}

	// 5. return response 201
	response := VisitorsPOSTResponseParameters{}
	response.Populate(createdVisitor)
	response.VisitorAccessToken = visitorAccessToken
	jsonutils.WriteJSON(w, http.StatusCreated, response)
}

func (cfg *ApiConfig) HandlerPutVisitorsByID(w http.ResponseWriter, r *http.Request) { // PUT /api/visitors/{visitor_id}
	// 1. Read endpoint URI for visitor ID, JWT for accessing user and authenticate based on either.
	visitorID, err := auth.VisitorsByID(w, r, cfg, cfg.DB)
	if err != nil {
		return
	}

	// 2. PUT request
	decoder := json.NewDecoder(r.Body)
	request := VisitorsPutRequestParameters{}
	err = decoder.Decode(&request)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "JSON formatting invalid")
		return
	}

	// 3. validate request? Purpose mainly.
	_, err = cfg.DB.GetPurposesByID(r.Context(), request.PurposeID)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, http.StatusNotFound, err, "purpose not found in database")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetPurposesById)")
		return
	}

	// 4. run query
	queryParams := database.SetVisitorByIDParams{
		ID:        visitorID,
		Name:      strutils.InitNullString(request.Name),
		PurposeID: request.PurposeID,
		Status:    request.Status,
	}
	updatedVisitor, err := cfg.DB.SetVisitorByID(r.Context(), queryParams)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, http.StatusNotFound, err, "updated visitor does not exist in database")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (SetVisitorByID)")
		return
	}

	// 6. write response
	response := VisitorsResponseParameters{}
	response.Populate(updatedVisitor)

	jsonutils.WriteJSON(w, http.StatusOK, response)

}

func (cfg *ApiConfig) HandlerGetVisitors(w http.ResponseWriter, r *http.Request) { // GET /api/visitors
	// 1. read request: JWT
	accessingUser, err := auth.UserFromHeader(w, r, cfg, cfg.DB)
	if err == auth.ErrWrongUserType {
		jsonutils.WriteError(w, http.StatusForbidden, err, "not logged in as user") // this is auth
		return
	} else if err != nil {
		return // auth.UserFromHeader() already calls jsonutils.WriteError() if something is wrong or the usertype isnt "user"
	} else if !accessingUser.IsActive {
		jsonutils.WriteError(w, http.StatusForbidden, err, "logged in user is not active") // when would this even happen?
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

	// 2.2 purpose name to purpose ID
	purpose, err := cfg.DB.GetPurposesByName(r.Context(), queryPurpose)
	if err == sql.ErrNoRows && queryPurpose != "" { // of course norows is not a problem if querypurpose is empty to begin with
		jsonutils.WriteError(w, http.StatusNotFound, err, "purpose not found in database")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetPurposesByName)")
		return
	}

	// 3. query database
	switch {
	case queryPurpose != "" && queryStatus != "": // both query parameters entered
		queryParams := database.GetVisitorsByPurposeStatusParams{
			PurposeID: purpose.ID,
			Status:    status,
		}
		visitors, err = cfg.DB.GetVisitorsByPurposeStatus(r.Context(), queryParams)
		if err == sql.ErrNoRows {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no visitors found in database for purpose "+queryPurpose+" and status "+queryStatus)
			return
		} else if err != nil {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetVisitorsByPurposeStatus)")
			return
		}

	case queryPurpose != "" && queryStatus == "": // only purpose query parameter entered
		visitors, err = cfg.DB.GetVisitorsByPurpose(r.Context(), purpose.ID)
		if err == sql.ErrNoRows {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no visitors found in databae for purpose "+queryPurpose)
			return
		} else if err != nil {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetVisitorsByPurpose)")
			return
		}

	case queryPurpose == "" && queryStatus != "": // only status query parameter entered
		visitors, err = cfg.DB.GetVisitorsByStatus(r.Context(), status)
		if err == sql.ErrNoRows {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no viistors found in databae for status "+queryStatus)
			return
		} else if err != nil {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetVisitorsByStatus)")
			return
		}

	case queryPurpose == "" && queryStatus == "": // neither query parameter entered
		visitors, err = cfg.DB.GetVisitors(r.Context())
		if err == sql.ErrNoRows {
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

func (cfg *ApiConfig) HandlerGetVisitorsByID(w http.ResponseWriter, r *http.Request) { // GET /api/visitors/{visitor_id}
	// 1. get visitor ID from endpoint
	// 2. read request: JWT
	// 3. authenticate: either for visitor with matching ID or user (both from JWT in 2)
	visitorID, err := auth.VisitorsByID(w, r, cfg, cfg.DB)
	if err != nil {
		return // visitorsbyID already handles all the jsonutils.WriteError() requirements as well as the authentication. No error means authentication is fine.
	}

	// 4. run query
	visitor, err := cfg.DB.GetVisitorByID(r.Context(), visitorID)
	if err == sql.ErrNoRows {
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

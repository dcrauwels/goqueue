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

type VisitorsPostRequestParameters struct {
	Name    string `json:"name"`
	Purpose string `json:"purpose"`
}

type VisitorsPutRequestParameters struct {
	Name    string `json:"name"`
	Purpose string `json:"purpose"`
	Status  int32  `json:"status"`
}

type VisitorsResponseParameters struct {
	ID           uuid.UUID      `json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	WaitingSince time.Time      `json:"waiting_since"`
	Name         sql.NullString `json:"name"`
	Purpose      string         `json:"purpose"`
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
	vrp.Purpose = v.Purpose
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
		jsonutils.WriteError(w, 400, err, "JSON formatting invalid")
		return
	}

	// 2. check purpose for validity NYI
	// these should be user defined, meaning a new migration is needed to:
	// a. CREATE TABLE purposes
	// b. UPDATE TABLE visitors ADD CONSTRAINT fk_purpose FOREIGN KEY purpose REFERENCES purposes (id)
	// not sure on b honestly. May need to down migrate

	// 3. query DB: CreateVisitor
	queryParams := database.CreateVisitorParams{
		Name:    strutils.InitNullString(request.Name), // name is currently nullable.
		Purpose: request.Purpose,
	}
	createdVisitor, err := cfg.DB.CreateVisitor(r.Context(), queryParams)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "could not query database to create visitor")
		return
	}

	// 4. make visitor access token
	visitorAccessToken, err := auth.MakeJWT(createdVisitor.ID, "visitor", cfg.Secret, 120)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "could not create access token")
		return
	}

	// 5. return response 201
	response := VisitorsPOSTResponseParameters{}
	response.Populate(createdVisitor)
	response.VisitorAccessToken = visitorAccessToken
	jsonutils.WriteJSON(w, 201, response)
}

func (cfg *ApiConfig) HandlerPutVisitorsByID(w http.ResponseWriter, r *http.Request) { // PUT /api/visitors/{visitor_id}
	// 1. Read endpoint URI for visitor ID, JWT for accessing user and authenticate based on either.
	visitorID, err := cfg.VisitorsById(w, r)
	if err != nil {
		return
	}

	// 2. PUT request
	decoder := json.NewDecoder(r.Body)
	request := VisitorsPutRequestParameters{}
	err = decoder.Decode(&request)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "JSON formatting invalid")
		return
	}

	// 3. validate request? Purpose mainly. NYI

	// 4. run query
	queryParams := database.SetVisitorByIDParams{
		ID:      visitorID,
		Name:    strutils.InitNullString(request.Name),
		Purpose: request.Purpose,
		Status:  request.Status,
	}
	updatedVisitor, err := cfg.DB.SetVisitorByID(r.Context(), queryParams)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "updated visitor does not exist in database")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database (SetVisitorByID)")
		return
	}

	// 6. write response
	response := VisitorsResponseParameters{}
	response.Populate(updatedVisitor)

	jsonutils.WriteJSON(w, 200, response)

}

func (cfg *ApiConfig) HandlerGetVisitors(w http.ResponseWriter, r *http.Request) { // GET /api/visitors
	// 1. read request: JWT
	accessingUser, err := auth.UserFromHeader(w, r, cfg, cfg.DB)
	if err == auth.ErrWrongUserType {
		jsonutils.WriteError(w, 403, err, "not logged in as user") // this is auth
		return
	} else if err != nil {
		return // auth.UserFromHeader() already calls jsonutils.WriteError() if something is wrong or the usertype isnt "user"
	} else if !accessingUser.IsActive {
		jsonutils.WriteError(w, 403, err, "logged in user is not active") // when would this even happen?
		return
	}

	// 2. validate request: purpose NYI

	// 3. run query
	visitors, err := cfg.DB.GetVisitors(r.Context())
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "no visitors found in database")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return
	}

	// 4. write response
	response := make([]VisitorsResponseParameters, len(visitors))
	for i, u := range visitors {
		response[i].Populate(u)
	}
}

func (cfg *ApiConfig) HandlerGetVisitorsByID(w http.ResponseWriter, r *http.Request) { // GET /api/visitors/{visitor_id}
	// 1. get visitor ID from endpoint
	// 2. read request: JWT
	// 3. authenticate: either for visitor with matching ID or user (both from JWT in 2)
	visitorID, err := cfg.VisitorsById(w, r)
	if err != nil {
		return // visitorsbyID already handles all the jsonutils.WriteError() requirements and handles the authentication.
	}

	// 4. run query
	visitor, err := cfg.DB.GetVisitorByID(r.Context(), visitorID)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "visitor not found in database")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database (GetVisitorByID)")
		return
	}

	// 5. write response
	response := VisitorsResponseParameters{}
	response.Populate(visitor)
	jsonutils.WriteJSON(w, 200, response)

}

func (cfg *ApiConfig) VisitorsById(w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	// boilerplate for GET and PUT /api/visitors/{visitor_id}
	// not sure the second and third return values (accessingID and userType)  are really needed
	// 1. read visitor ID from endpoint URI
	pv := r.PathValue("visitor_id")
	visitorID, err := uuid.Parse(pv)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "endpoint is not a valid ID")
		return visitorID, err
	}

	// 2. read request data: JWT
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		jsonutils.WriteError(w, 401, err, "no authorization field in request header")
		return visitorID, err
	}

	accessingID, userType, err := auth.ValidateJWT(accessToken, cfg.GetSecret())
	if err != nil {
		jsonutils.WriteError(w, 403, err, "access token invalid") // is a 403 even the right response code here?
		return visitorID, err
	}

	// 3. authenticate: either for visitor with matching ID or user (both from JWT in 2)
	if userType == "user" { // auth for user
		accessingParty, err := cfg.DB.GetUserByID(r.Context(), accessingID)
		if err == sql.ErrNoRows {
			jsonutils.WriteError(w, 404, err, "accessing user does not exist in database")
			return visitorID, err
		} else if err != nil {
			jsonutils.WriteError(w, 500, err, "error querying database (GetUserByID)")
			return visitorID, err
		} else if !accessingParty.IsActive {
			jsonutils.WriteError(w, 403, auth.ErrWrongUserType, "accessing user account is inactive")
			return visitorID, auth.ErrWrongUserType
		}
	} else if userType == "visitor" { // auth for visitor
		accessingParty, err := cfg.DB.GetVisitorByID(r.Context(), accessingID)
		if err == sql.ErrNoRows {
			jsonutils.WriteError(w, 404, err, "accessing visitor does not exist in database")
			return visitorID, err
		} else if err != nil {
			jsonutils.WriteError(w, 500, err, "error querying database (GetVisitorByID)")
			return visitorID, err
		} else if accessingParty.ID != visitorID { // so the visitor is trying to edit a different visitor. not allowed
			jsonutils.WriteError(w, 403, auth.ErrVisitorMismatch, "accessing visitor is not requested visitor")
			return visitorID, auth.ErrVisitorMismatch
		} else if accessingParty.ID != accessingID { // sanity check
			jsonutils.WriteError(w, 500, auth.ErrVisitorMismatch, "accessing visitor is not corresponding to database")
			return visitorID, auth.ErrVisitorMismatch
		}
	} else { // in case usertype is neither "user" nor "visitor"
		jsonutils.WriteError(w, 400, auth.ErrWrongUserType, "incorrect usertype in JWT")
		return visitorID, auth.ErrWrongUserType
	}

	return visitorID, nil
}

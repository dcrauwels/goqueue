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

/*type VisitorsResponseParameters struct {
	PublicID          string         `json:"public_id"`
	WaitingSince      time.Time      `json:"waiting_since"`
	Name              sql.NullString `json:"name"`
	PurposePublicID   string         `json:"purpose_public_id"`
	Status            int32          `json:"status"`
	DailyTicketNumber int32          `json:"daily_ticket_number"`
	PurposeName       string         `json:"purpose_name"`
}

func (vrp *VisitorsResponseParameters) Populate(v database.CreateVisitorRow) {
	vrp.PublicID = v.PublicID
	vrp.WaitingSince = v.WaitingSince
	vrp.Name = v.Name
	vrp.PurposePublicID = v.PurposePublicID
	vrp.PurposeName = v.PurposeName
	vrp.Status = v.Status
	vrp.DailyTicketNumber = v.DailyTicketNumber
}*/

type VisitorStatus int32

const (
	StatusCancelled     VisitorStatus = iota // 0
	StatusWaiting                            // 1
	StatusCalled                             // 2
	StatusInService                          // 3
	StatusCompleted                          // 4
	StatusNoShow                             // 5
	StatusAutoCompleted                      // 6
)

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

	// 5. query DB: CreateVisitor
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

	// 6. send message to broker
	cfg.Broker.Notifier <- []byte("refresh_queue")

	// 7. return response 201
	jsonutils.WriteJSON(w, http.StatusCreated, createdVisitor)
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
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB) // I don't need information about the user itself, just whether a user ID is present in the request context.
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access PUT /api/visitors")
		return
	} else if !accessingUser.IsActive {
		jsonutils.WriteError(w, http.StatusForbidden, auth.ErrUserInactive, "accessing user account is inactive")
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

	// 4. run query
	queryParams := database.SetVisitorByPublicIDParams{
		PublicID:        pvid,
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
	jsonutils.WriteJSON(w, http.StatusOK, updatedVisitor)
}

func (cfg *ApiConfig) HandlerGetVisitors(w http.ResponseWriter, r *http.Request) { // GET /api/visitors
	// only accessible to logged in users
	// 1. get user authentication from request context
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB) // not interested in actual information about the user
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access GET /api/visitors")
		return
	} else if !accessingUser.IsActive {
		jsonutils.WriteError(w, http.StatusForbidden, auth.ErrUserInactive, "accessing user account is inactive")
		return
	}

	// 2. check for query parameters (purpose, status)
	q := r.URL.Query()
	params := database.ListVisitorsParams{
		PurposePublicID: strutils.QueryParameterToNullString(q.Get("purpose")),
	}

	// 2.1 status as string to status as int32
	status, err := strutils.QueryParameterToNullInt(q.Get("status"))
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "query parameter 'status' takes integers")
		return
	}
	params.Status = status

	// 2.3 start and end dates
	var t sql.NullTime
	t, err = strutils.QueryParameterToNullTime(q.Get("start_date"))
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "query parameter 'start_date' takes ISO 8601 format (YYYY-MM-DD)")
		return
	}
	params.StartDate = t

	t, err = strutils.QueryParameterToNullTime(q.Get("end_date"))
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "query parameter 'end_date' takes ISO 8601 format (YYYY-MM-DD)")
		return
	}
	params.EndDate = t

	// 3. query database
	visitors, err := cfg.DB.ListVisitors(r.Context(), params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no visitors found under specified query parameters")
			return
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (ListVisitors in HandlerGetVisitors)")
			return
		}
	}

	// 4. write response
	jsonutils.WriteJSON(w, http.StatusOK, visitors)
}

func (cfg *ApiConfig) HandlerGetVisitorsByPublicID(w http.ResponseWriter, r *http.Request) { // GET /api/visitors/{visitor_public_id}
	// 1. get visitor ID from endpoint
	pvid, err := strutils.GetPublicIDFromPathValue("visitor_public_id", cfg.PublicIDLength, r)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "incorrect path value length")
		return
	}

	// 2. run query
	visitor, err := cfg.DB.GetVisitorsByPublicID(r.Context(), pvid)
	if errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusNotFound, err, "visitor not found in database")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetVisitorByID)")
		return
	}

	// 3. write response
	jsonutils.WriteJSON(w, http.StatusOK, visitor)

}

func (cfg *ApiConfig) HandlerGetQueue(w http.ResponseWriter, r *http.Request) { // GET /api/visitors/queue
	// 1. auth
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication is required for this endpoint (GET /api/queue)")
		return
	} else if !accessingUser.IsActive {
		jsonutils.WriteError(w, http.StatusForbidden, auth.ErrUserInactive, "accessing user account is inactive")
		return
	}

	// 2. run query
	visitors, err := cfg.DB.GetQueue(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no rows found")
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying server (GetQueue in HandlerGetQueue)")
		}
		return
	}

	// 3. write response
	jsonutils.WriteJSON(w, http.StatusOK, visitors)
}

func (cfg *ApiConfig) HandlerCallNextVisitor(w http.ResponseWriter, r *http.Request) { // POST /api/visitors/call-next
	// 1. check auth
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication is required for this endpoint")
		return
	} else if !accessingUser.IsActive {
		jsonutils.WriteError(w, http.StatusForbidden, auth.ErrUserInactive, "accessing user account is inactive")
		return
	} else if !accessingUser.DeskPublicID.Valid {
		jsonutils.WriteError(w, http.StatusForbidden, auth.ErrUserWithoutDesk, "accessing user is not at a desk")
		return
	}

	// 2. check for active visitors on accessing user
	// 2.1 query DB for servicelogs by user public ID
	serviceLogs, err := cfg.DB.GetActiveServiceLogsByUserID(r.Context(), accessingUser.PublicID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { // getting errnorows is expected in case no active service logs are available for this user
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetActiveServiceLogsByUserID in HandlerCallNextVisitor)")
			return
		}
	}

	// 2.2 deactivate servicelogs and set visitors to StatusCompleted
	if len(serviceLogs) != 0 {
		for _, sl := range serviceLogs {
			// 2.2.1 set visitor status
			oldVisitorParams := database.SetVisitorStatusByPublicIDParams{
				PublicID: sl.VisitorPublicID,
				Status:   int32(StatusCompleted),
			}
			_, err := cfg.DB.SetVisitorStatusByPublicID(r.Context(), oldVisitorParams) // updated old visitor is not returned
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					jsonutils.WriteError(w, http.StatusNotFound, err, "visitor public ID in active service log for accessing user does not match any existing visitors")
					return
				} else {
					jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (SetVisitorStatusByPublicID in HandlerCallNextVisitor)")
					return
				}
			}

			// 2.2.2 set servicelog inactive
			oldServiceLogParams := database.SetServiceLogsIsActiveByPublicIDParams{
				PublicID: sl.PublicID,
				IsActive: false,
			}
			_, err = cfg.DB.SetServiceLogsIsActiveByPublicID(r.Context(), oldServiceLogParams) // updated old servicelog is not returned
			if err != nil {                                                                    // no need to check for sql.ErrNoRows -- this would be extremely bizarre, as this entire loop runs over the very servicelog we are querying here
				jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (SetServiceLogsIsActiveByPublicID in HandlerCallNextVisitor)")
				return
			}

		}
	}

	// 3. get next waiting visitor. Note that this updates  the visitor in the SQL query already
	calledVisitor, err := cfg.DB.CallNextWaitingVisitor(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteJSON(w, http.StatusOK, nil)
			return
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (CallNextVisitor in HandlerCallNextVisitor)")
			return
		}
	}

	// 4. insert servicelog
	slQueryParams := database.CreateServiceLogsParams{
		PublicID:        cfg.PublicIDGenerator(),
		VisitorPublicID: calledVisitor.PublicID,
		UserPublicID:    accessingUser.PublicID,
		DeskPublicID:    accessingUser.DeskPublicID.String,
	}
	createdServiceLog, err := cfg.DB.CreateServiceLogs(r.Context(), slQueryParams)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (CreateServiceLogs in HandlerCallNextVisitor)")
		return
	}

	// 5. notify broker
	cfg.Broker.Notifier <- []byte("refresh_queue")

	// 6. write response
	accessingUserDesk, err := cfg.DB.GetDesksByPublicID(r.Context(), accessingUser.DeskPublicID.String)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetDesksByPublicID in HandlerCallNextVisitor)")
	}

	type CallNextResponse struct {
		ServicelogPublicID string    `json:"servicelog_public_id"`
		CalledAt           time.Time `json:"called_at"`
		IsActive           bool      `json:"is_active"`
		Visitor            struct {
			VisitorPublicID   string         `json:"visitor_public_id"`
			VisitorName       sql.NullString `json:"visitor_name"`
			DailyTicketNumber int32          `json:"daily_ticket_number"`
			Status            int32          `json:"status"`
			WaitingSince      time.Time      `json:"waiting_since"`
			PurposePublicID   string         `json:"purpose_public_id"`
			PurposeName       string         `json:"purpose_name"`
		}
		Desk struct {
			DeskPublicID string `json:"desk_public_id"`
			DeskName     string `json:"desk_name"`
		}
	}

	response := CallNextResponse{
		ServicelogPublicID: createdServiceLog.PublicID,
		CalledAt:           createdServiceLog.CalledAt,
		IsActive:           createdServiceLog.IsActive,
	}
	response.Visitor.VisitorPublicID = calledVisitor.PublicID
	response.Visitor.VisitorName = calledVisitor.Name
	response.Visitor.DailyTicketNumber = calledVisitor.DailyTicketNumber
	response.Visitor.Status = calledVisitor.Status
	response.Visitor.WaitingSince = calledVisitor.WaitingSince
	response.Visitor.PurposePublicID = calledVisitor.PurposePublicID
	response.Visitor.PurposeName = calledVisitor.PurposeName
	response.Desk.DeskPublicID = accessingUserDesk.PublicID
	response.Desk.DeskName = accessingUserDesk.Name

	jsonutils.WriteJSON(w, http.StatusOK, response)
}

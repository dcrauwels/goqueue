package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/dcrauwels/goqueue/strutils"
	"github.com/google/uuid"
)

type VisitorsRequestParameters struct {
	Name    string `json:"name"`
	Purpose string `json:"purpose"`
}

type VisitorsResponseParameters struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	WaitingSince time.Time `json:"waiting_since"`
	Name         string    `json:"name"`
	Purpose      string    `json:"purpose"`
	Status       int32     `json:"status"`
}

func (cfg *ApiConfig) HandlerPostVisitors(w http.ResponseWriter, r *http.Request) {
	// function for sending a POST request to register a single visitor from scratch
	// in context the visitor accesses a website, enters his name and purpose and gets a number
	//
	// 1. get request data: name, purpose
	decoder := json.NewDecoder(r.Body)
	reqParams := VisitorsRequestParameters{}
	err := decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "JSON formatting invalid")
		return
	}

	// 2. make visitor refresh token NYI THIS IS BIG

	// 3. check purpose for validity NYI

	// 4. query DB: CreateVisitor
	queryParams := database.CreateVisitorParams{
		Name:    strutils.InitNullString(reqParams.Name), // name is currently nullable.
		Purpose: reqParams.Purpose,
	}
	createdVisitor, err := cfg.DB.CreateVisitor(r.Context(), queryParams)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "could not query database to create visitor")
		return
	}

	// 5. return response 201
	responseParams := VisitorsResponseParameters{
		ID:           createdVisitor.ID,
		CreatedAt:    createdVisitor.CreatedAt,
		UpdatedAt:    createdVisitor.UpdatedAt,
		WaitingSince: createdVisitor.WaitingSince,
		Name:         createdVisitor.Name.String,
		Purpose:      createdVisitor.Purpose,
		Status:       createdVisitor.Status,
	}
	jsonutils.WriteJSON(w, 201, responseParams)
}

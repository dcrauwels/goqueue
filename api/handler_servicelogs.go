package api

import (
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/google/uuid"
)

type ServicelogsRequestParameters struct {
	VisitorID uuid.UUID
	UserID    uuid.UUID
	DeskID    uuid.UUID
}

type ServicelogsResponseParameters struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	VisitorID uuid.UUID
	UserID    uuid.UUID
	DeskID    uuid.UUID
	CalledAt  time.Time
	IsActive  bool
}

func (slrp *ServicelogsResponseParameters) Populate(sl database.ServiceLog) {
	slrp.ID = sl.ID
	slrp.CreatedAt = sl.CreatedAt

}

type ServiceLog struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	VisitorID uuid.UUID
	UserID    uuid.UUID
	DeskID    uuid.UUID
	CalledAt  time.Time
	IsActive  bool
}

// POST /api/servicelogs (user only)
func (cfg *ApiConfig) HandlerPostServicelogs(w http.ResponseWriter, r *http.Request) {
	// 1. Check auth
	// 2. Read request
	// 3. Query DB
	// 4. Handle errors
	// 5. Send response
}

// PUT /api/servicelogs/{servicelog_id} (user only)
func (cfg *ApiConfig) HandlerPutServicelogsByID(w http.ResponseWriter, r *http.Request) {
}

// GET /api/servicelogs (user only)
func (cfg *ApiConfig) HandlerGetServicelogs(w http.ResponseWriter, r *http.Request) {
}

// GET /api/servicelogs/{visitor_id} not convinced this is needed
//func (cfg *ApiConfig) HandlerGetServicelogsByVisitorID(w http.ResponseWriter, r *http.Request) {
//}

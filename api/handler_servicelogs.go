package api

import (
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/google/uuid"
)

type ServicelogsPOSTRequestParameters struct {
	VisitorID uuid.UUID `json:"visitor_id"`
	UserID    uuid.UUID `json:"user_id"`
	DeskID    uuid.UUID `json:"desk_id"`
}

type ServicelogsResponseParameters struct {
	ID        uuid.UUID `json:"id"`
	PublicID  string    `json:"public_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	VisitorID uuid.UUID `json:"visitor_id"`
	UserID    uuid.UUID `json:"user_id"`
	DeskID    uuid.UUID `json:"desk_id"`
	CalledAt  time.Time `json:"called_at"`
	IsActive  bool      `json:"is_active"`
}

func (slrp *ServicelogsResponseParameters) Populate(sl database.ServiceLog) {
	slrp.ID = sl.ID
	slrp.CreatedAt = sl.CreatedAt

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

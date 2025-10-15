package api

import (
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

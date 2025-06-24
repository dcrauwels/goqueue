package auth

import (
	"context"
	"errors"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/google/uuid"
)

var ErrWrongUserType = errors.New("usertype supplied in JWT is not valid")

var ErrVisitorMismatch = errors.New("accessing visitor is not visitor identified in endpoint URI")

var ErrUserInactive = errors.New("user account is inactive")

type configReader interface {
	GetSecret() string
}

type databaseQueryer interface {
	GetUserByID(context.Context, uuid.UUID) (database.User, error)
	GetVisitorByID(context.Context, uuid.UUID) (database.Visitor, error)
}

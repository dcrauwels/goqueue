package api

import (
	"context"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/google/uuid"
)

type ApiConfig struct {
	DB     *database.Queries
	Secret string
}

func (cfg *ApiConfig) GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error) {
	return cfg.DB.GetUserByID(ctx, id)
}

func (cfg *ApiConfig) GetSecret() string {
	return cfg.Secret
}

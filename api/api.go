package api

import (
	"context"
	"net/http"

	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/google/uuid"
)

type ApiConfig struct {
	DB     *database.Queries
	Secret string
	Env    string
}

func (cfg *ApiConfig) GetUserByID(ctx context.Context, id uuid.UUID) (database.User, error) {
	return cfg.DB.GetUserByID(ctx, id)
}

func (cfg ApiConfig) GetSecret() string {
	return cfg.Secret
}

func (cfg ApiConfig) GetEnv() string {
	return cfg.Env
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	cfg.HandlerPostUsers(w, r)

}

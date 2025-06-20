package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dcrauwels/goqueue/auth"
	"github.com/dcrauwels/goqueue/internal/database"
	"github.com/dcrauwels/goqueue/jsonutils"
	"github.com/google/uuid"
)

type responseParameters struct {
	ID               uuid.UUID `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Email            string    `json:"email"`
	IsAdmin          bool      `json:"is_admin"`
	IsActive         bool      `json:"is_active"`
	UserAccessToken  string    `json:"user_access_token"`
	UserRefreshToken string    `json:"user_refresh_token"`
}

func (cfg *ApiConfig) HandlerLoginUser(w http.ResponseWriter, r *http.Request) {
	// for authenticating USERS, not VISITORS
	// 1. get request content (email: string, password: string)
	decoder := json.NewDecoder(r.Body)
	reqParams := UsersRequestParameters{}
	err := decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "JSON formatting invalid")
		return
	}

	// 2. validate login credentials
	user, err := cfg.DB.GetUserByEmail(r.Context(), reqParams.Email)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "email or password incorrect")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return
	}
	err = auth.CheckPasswordHash(user.HashedPassword, reqParams.Password)
	if err != nil {
		jsonutils.WriteError(w, 404, err, "email or password incorrect")
		return
	}
	// 2.5 check if refresh token already exists
	_, err = cfg.DB.GetRefreshTokensByUserID(r.Context(), user.ID)
	if err == nil {
		jsonutils.WriteError(w, 403, err, "user already logged in, use /api/refresh endpoint instead")
		return
	} else if err != nil && err != sql.ErrNoRows { // tautological for clarity
		jsonutils.WriteError(w, 500, err, "error querying database")
		return
	}

	// 3. generate access token
	userAccessToken, err := auth.MakeJWT(user.ID, cfg.Secret, 60)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "error creating access token")
		return
	}

	// 4. query for refresh token
	newRefreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		jsonutils.WriteError(w, 500, err, "error creating refresh token")
		return
	}
	queryParams := database.CreateRefreshTokenParams{
		Token:     newRefreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 14), // 14 days
	}
	_, err = cfg.DB.CreateRefreshToken(r.Context(), queryParams)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return
	}

	// 5. return access token to user

	respParams := responseParameters{
		ID:               user.ID,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		Email:            user.Email,
		IsAdmin:          user.IsAdmin,
		IsActive:         user.IsActive,
		UserAccessToken:  userAccessToken,
		UserRefreshToken: newRefreshToken,
	}

	jsonutils.WriteJSON(w, 200, respParams)

}

func (cfg *ApiConfig) HandlerRefreshUser(w http.ResponseWriter, r *http.Request) {
	// for getting USERS a new access token based on a valid refresh token
	// 1. get request content (refresh token & username combo)
	type requestParameters struct {
		RefreshToken string    `json:"refresh_token"`
		UserID       uuid.UUID `json:"user_id"`
	}
	reqParams := requestParameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, 400, err, "invalid JSON request structure")
		return
	}

	// 2. validate refresh token through DB query
	fullRefreshToken, err := cfg.DB.GetRefreshTokenByToken(r.Context(), reqParams.RefreshToken)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "refresh token not found")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return
	}
	if reqParams.UserID != fullRefreshToken.UserID || fullRefreshToken.ExpiresAt.After(time.Now()) {
		jsonutils.WriteError(w, 403, err, "refresh token invalid")
		return
	}

	// 3. generate access token
	userAccessToken, err := auth.MakeJWT(reqParams.UserID, cfg.Secret, 60)
	if err != nil {
		jsonutils.WriteError(w, 500, err, "error creating access token")
		return
	}

	// 4. return access token to user
	fullUser, err := cfg.DB.GetUserByID(r.Context(), reqParams.UserID)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 404, err, "user does not exist any more")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return
	}
	respParams := responseParameters{
		ID:               fullUser.ID,
		CreatedAt:        fullUser.CreatedAt,
		UpdatedAt:        fullUser.UpdatedAt,
		Email:            fullUser.Email,
		IsAdmin:          fullUser.IsAdmin,
		IsActive:         fullUser.IsActive,
		UserAccessToken:  userAccessToken,
		UserRefreshToken: fullRefreshToken.Token,
	}
	jsonutils.WriteJSON(w, 200, respParams)
}

func (cfg *ApiConfig) HandlerLogoutUser(w http.ResponseWriter, r *http.Request) {
	// for revoking USER refresh token
	// 1. get user from token (auth.fromheader)
	user, err := auth.UserFromHeader(w, r, cfg)
	if err != nil {
		return
	}
	// 2. query to revoke refresh token
	_, err = cfg.DB.RevokeRefreshTokenByUserID(r.Context(), user.ID)
	if err == sql.ErrNoRows {
		jsonutils.WriteError(w, 403, err, "no refresh token found for this user")
		return
	} else if err != nil {
		jsonutils.WriteError(w, 500, err, "error querying database")
		return
	}

	// 3. send response with empty access token
	respParams := responseParameters{
		ID:               user.ID,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		Email:            user.Email,
		IsAdmin:          user.IsAdmin,
		IsActive:         user.IsActive,
		UserAccessToken:  "",
		UserRefreshToken: "",
	}
	jsonutils.WriteJSON(w, 200, respParams)

}

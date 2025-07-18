package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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

type refreshTokenResponseParams struct {
	Token     string       `json:"token"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	UserID    uuid.UUID    `json:"user_id"`
	ExpiresAt time.Time    `json:"expires_at"`
	RevokedAt sql.NullTime `json:"revoked_at"`
}

func (cfg *ApiConfig) HandlerGetRefreshTokens(w http.ResponseWriter, r *http.Request) { // GET /api/refresh
	// 1. authenticate: only dev environment
	if cfg.Env != "dev" {
		jsonutils.WriteError(w, http.StatusMethodNotAllowed, fmt.Errorf("user tried to access GET /api/refresh from incorrect environment"), "GET not allowed for this endpoint")
		return
	}

	// 2. get query parameters (user)
	queryParameters := r.URL.Query()
	queryUser := queryParameters.Get("user")

	// 3. run query
	var refreshTokens []database.RefreshToken
	var err error
	if queryUser != "" {
		userID, err := uuid.Parse(queryUser)
		if err != nil {
			jsonutils.WriteError(w, http.StatusBadRequest, err, "query parameter (?user=) is not a valid user ID")
			return
		}
		refreshTokens, err = cfg.DB.GetRefreshTokensByUserID(r.Context(), userID) // note that this returns an empty slice, not a sql.ErrNoRows error in case no rows are found!
	} else {
		refreshTokens, err = cfg.DB.GetRefreshTokens(r.Context())
	}

	if errors.Is(err, sql.ErrNoRows) || len(refreshTokens) == 0 {
		jsonutils.WriteError(w, http.StatusNotFound, err, "no rows found when querying database (GetRefreshTokens in HandlerGetRefreshTokens)")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetRefreshRokens in HandlerGetRefreshTokens)")
		return
	}

	// 3. write response
	response := make([]refreshTokenResponseParams, len(refreshTokens))
	for i, u := range refreshTokens {
		response[i].Token = u.Token
		response[i].CreatedAt = u.CreatedAt
		response[i].UpdatedAt = u.UpdatedAt
		response[i].UserID = u.UserID
		response[i].ExpiresAt = u.ExpiresAt
		response[i].RevokedAt = u.RevokedAt
	}
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

func (cfg *ApiConfig) HandlerLoginUser(w http.ResponseWriter, r *http.Request) {
	// for authenticating USERS, not VISITORS
	// 1. get request content (email: string, password: string)
	decoder := json.NewDecoder(r.Body)
	reqParams := UsersRequestParameters{}
	err := decoder.Decode(&reqParams)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "JSON formatting invalid")
		return
	}

	// 2. validate login credentials
	user, err := cfg.DB.GetUserByEmail(r.Context(), reqParams.Email)
	if errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusNotFound, err, "email or password incorrect")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetUserByEmail in HandlerLoginUser)")
		return
	}
	err = auth.CheckPasswordHash(user.HashedPassword, reqParams.Password)
	if err != nil {
		jsonutils.WriteError(w, http.StatusNotFound, err, "email or password incorrect")
		return
	}
	// 2.5 check if refresh token already exists
	_, err = cfg.DB.GetRefreshTokensByUserID(r.Context(), user.ID)
	if err == nil {
		jsonutils.WriteError(w, http.StatusSeeOther, err, "user already logged in, use /api/refresh endpoint instead")
		http.Redirect(w, r, "/api/refresh", http.StatusSeeOther)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetRefreshTokensByUserID in HandlerLoginUser)")
		return
	}

	// 3. generate access token
	userAccessToken, err := auth.MakeJWT(user.ID, "user", cfg.Secret, 60)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error creating access token (in HandlerLoginUser)")
		return
	}

	// 4. query for refresh token
	newRefreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error creating refresh token (in HandlerLoginUser)")
		return
	}
	queryParams := database.CreateRefreshTokenParams{
		Token:     newRefreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 14), // 14 days
	}
	_, err = cfg.DB.CreateRefreshToken(r.Context(), queryParams)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (CreateRefreshToken in HandlerLoginUser)")
		return
	}

	// 5. write cookies
	auth.SetAuthCookies(w, userAccessToken, newRefreshToken, "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)

	// 6. return access token to user
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

	jsonutils.WriteJSON(w, http.StatusOK, respParams)

}

func (cfg *ApiConfig) HandlerRefreshUser(w http.ResponseWriter, r *http.Request) { // POST /api/refresh
	// for getting USERS a new access token based on a valid refresh token
	// many problems
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

	// 2.5 rotate refresh token (revoke old token, return new token) NYI

	// 3. generate access token
	userAccessToken, err := auth.MakeJWT(reqParams.UserID, "user", cfg.Secret, 60)
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
	// 1. get user from context
	user, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil { // error handling
		return
	}

	// 2. query to revoke refresh token
	_, err = cfg.DB.RevokeRefreshTokenByUserID(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, 403, err, "no refresh token found for this user")
			return
		} else {
			jsonutils.WriteError(w, 500, err, "error querying database")
			return
		}
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

func (cfg *ApiConfig) HandlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) { // POST /api/revoke NYI
	// 1. Get request data (token, user)
	// 2. validate: token matches user ID?
	// 3. query database cfg.DB.RevokeRefreshTokenByToken
	// 4. write response
}

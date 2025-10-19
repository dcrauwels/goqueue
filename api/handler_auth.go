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

type loginRequestParameters struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type responseParameters struct {
	ID               uuid.UUID `json:"id"`
	PublicID         string    `json:"public_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Email            string    `json:"email"`
	FullName         string    `json:"full_name"`
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

func (rp *refreshTokenResponseParams) Populate(token database.RefreshToken) {
	rp.Token = token.Token
	rp.CreatedAt = token.CreatedAt
	rp.UpdatedAt = token.UpdatedAt
	rp.UserID = token.UserID
	rp.ExpiresAt = token.ExpiresAt
	rp.RevokedAt = token.RevokedAt
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
		response[i].Populate(u)
	}
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

func (cfg *ApiConfig) HandlerLoginUser(w http.ResponseWriter, r *http.Request) {
	// for authenticating USERS, not VISITORS
	// 1. get request content (email: string, password: string)
	decoder := json.NewDecoder(r.Body)
	reqParams := loginRequestParameters{}
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
	// 2.1 check if refresh token already exists
	_, err = cfg.DB.GetRefreshTokensByUserID(r.Context(), user.ID)
	if err == nil {
		jsonutils.WriteError(w, http.StatusSeeOther, err, "user already logged in, use /api/refresh endpoint instead")
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
		PublicID:         user.PublicID,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		Email:            user.Email,
		FullName:         user.FullName,
		IsAdmin:          user.IsAdmin,
		IsActive:         user.IsActive,
		UserAccessToken:  userAccessToken,
		UserRefreshToken: newRefreshToken,
	}

	jsonutils.WriteJSON(w, http.StatusOK, respParams)

}

func (cfg *ApiConfig) HandlerRefreshUser(w http.ResponseWriter, r *http.Request) { // POST /api/refresh
	// for getting USERS a new access token based on a valid refresh token
	// 1. get user from context
	user, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil { // error handling
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access /api/logout")
		return
	}

	// 2. validate refresh token through DB query
	validRefreshTokens, err := cfg.DB.GetRefreshTokensByUserID(r.Context(), user.ID)
	if errors.Is(err, sql.ErrNoRows) {
		jsonutils.WriteError(w, http.StatusNotFound, err, "refresh token not found")
		return
	} else if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (GetRefreshTokensByUserID in HandlerRefreshUser)")
		return
	}
	// 2.1 if more than one valid refresh token is available: rotate fully (revoke all, return new token)
	if len(validRefreshTokens) != 1 {
		_, err = cfg.DB.RevokeRefreshTokenByUserID(r.Context(), user.ID)
		if err != nil {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (RevokeRefreshTokensByUserID in HandlerRefreshUsers)")
			return
		}
		jsonutils.WriteError(w, http.StatusUnauthorized, errors.New("auth: user has more than one valid refresh token"), "too many refresh tokens found")
		return
	}

	// 2.2 define single refresh token
	refreshToken := validRefreshTokens[0]

	// 2.3 check token expiration. UserID check should be superfluous (as the query should also test for this) but why not
	if user.ID != refreshToken.UserID || refreshToken.ExpiresAt.After(time.Now()) {
		_, err = cfg.DB.RevokeRefreshTokenByToken(r.Context(), refreshToken.Token)
		if err != nil {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database (RevokeRefreshTokenByToken in HandlerRefreshUsers)")
			return
		}
		jsonutils.WriteError(w, http.StatusUnauthorized, errors.New("auth: refresh token not valid"), "invalid refresh token provided in context")
		return
	}

	// 2.4 rotate refresh token
	refreshTokenCookie, err := r.Cookie("user_refresh_token")
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "user_refresh_token cookie not found")
		return
	}
	newRefreshToken, err := auth.RotateRefreshToken(cfg.DB, w, r, refreshTokenCookie)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error rotating refresh token (RotateRefreshToken in HandlerRefreshUser)")
		return
	}

	// 3. generate access token
	userAccessToken, err := auth.MakeJWT(user.ID, "user", cfg.Secret, cfg.AccessTokenDuration)
	if err != nil {
		jsonutils.WriteError(w, http.StatusInternalServerError, err, "error creating access token")
		return
	}

	// 4. return access token to user > Not sure this is correct
	respParams := responseParameters{
		ID:               user.ID,
		PublicID:         user.PublicID,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		Email:            user.Email,
		FullName:         user.FullName,
		IsAdmin:          user.IsAdmin,
		IsActive:         user.IsActive,
		UserAccessToken:  userAccessToken,
		UserRefreshToken: newRefreshToken.Token,
	}
	jsonutils.WriteJSON(w, http.StatusOK, respParams)
}

func (cfg *ApiConfig) HandlerLogoutUser(w http.ResponseWriter, r *http.Request) {
	// for revoking USER refresh token
	// 1. get user from context
	user, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil { // error handling
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access /api/logout")
		return
	}

	// 2. query to revoke refresh token
	_, err = cfg.DB.RevokeRefreshTokenByUserID(r.Context(), user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) { // highly unexpected state
			jsonutils.WriteError(w, http.StatusNotFound, err, "no refresh token found for this user")
			auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration) // so remove all cookies just in case
			return
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying database")
			return
		}
	}

	// 3. empty cookies
	auth.SetAuthCookies(w, "", "", "user", cfg.AccessTokenDuration, cfg.RefreshTokenDuration)
	// 3.1 and send response with empty access token
	respParams := responseParameters{
		ID:               user.ID,
		PublicID:         user.PublicID,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
		Email:            user.Email,
		FullName:         user.FullName,
		IsAdmin:          user.IsAdmin,
		IsActive:         user.IsActive,
		UserAccessToken:  "",
		UserRefreshToken: "",
	}
	jsonutils.WriteJSON(w, http.StatusOK, respParams)

}

func (cfg *ApiConfig) HandlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) { // POST /api/revoke/{user_id} NYI
	/*
		Function for revoking a single user's refresh token(s).
	*/

	// 1. Get userID from URI
	req := r.PathValue("user_id")
	userID, err := uuid.Parse(req)
	if err != nil {
		jsonutils.WriteError(w, http.StatusBadRequest, err, "endpoint is not a valid user ID")
		return
	}

	// 2. Authenticate from context: get user and check if admin (and if URI == accessing user, redirect to /api/logout)
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access POST /api/revoke")
		return
	} else if accessingUser.ID == userID { // in this case the user is sending a revoke request for themselves, which should be a logout instead
		jsonutils.WriteJSON(w, http.StatusBadRequest, "user is trying to revoke self - this is done through POST /api/logout")
		return
	} else if !accessingUser.IsAdmin {
		jsonutils.WriteJSON(w, http.StatusForbidden, "non-admin users are not allowed to send POST requests to /api/revoke")
		return
	}

	// 3. query database cfg.DB.RevokeRefreshTokenByUserID
	revokedTokens, err := cfg.DB.RevokeRefreshTokenByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no valid refresh tokens found for this user")
			return
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying datatabase (RevokeRefreshTokenByUserID in HandlerRevokeRefreshToken)")
			return
		}
	}

	// 4. write response
	response := make([]refreshTokenResponseParams, len(revokedTokens))
	for i, u := range revokedTokens {
		response[i].Populate(u)
	}
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

func (cfg *ApiConfig) HandlerRevokeAllRefreshTokens(w http.ResponseWriter, r *http.Request) { // POST /api/revoke
	/*
		Function for taking the rather nuclear option of revoking all refresh tokens. This means all users are instantly logged out.
		Like POST /api/revoke/{user_id} this should be restricted to admin type users only. Part of me wonders whether to have this at all.
	*/

	// 1. Authenticate from context: get user and check if admin
	accessingUser, err := auth.UserFromContext(w, r, cfg.DB)
	if err != nil {
		jsonutils.WriteError(w, http.StatusUnauthorized, err, "user authentication required to access POST /api/revoke")
		return
	} else if !accessingUser.IsAdmin {
		jsonutils.WriteError(w, http.StatusForbidden, err, "non-admin users are not allowed to send POST requests to /api/revoke")
		return
	}

	// 2. query database cfg.DB.RevokeRefreshTokens
	revokedTokens, err := cfg.DB.RevokeRefreshTokens(r.Context())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			jsonutils.WriteError(w, http.StatusNotFound, err, "no valid refresh tokens found")
			return
		} else {
			jsonutils.WriteError(w, http.StatusInternalServerError, err, "error querying datatabase (RevokeRefreshTokens in HandlerRevokeRefreshToken)")
			return
		}
	}

	// 3. write response
	response := make([]refreshTokenResponseParams, len(revokedTokens))
	for i, u := range revokedTokens {
		response[i].Populate(u)
	}
	jsonutils.WriteJSON(w, http.StatusOK, response)
}

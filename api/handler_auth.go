package api

import (
	"net/http"
)

func (cfg *ApiConfig) HandlerLoginUser(w http.ResponseWriter, r *http.Request) {
	// for authenticating USERS, not VISITORS
	// 1. get request content (email: string, password: string)
	// 2. validate login credentials
	// 3. generate access token
	// 4. query for refresh token
	// 5. return access token to user
}

func (cfg *ApiConfig) HandlerRefreshUser(w http.ResponseWriter, r *http.Request) {
	// for getting USERS a new access token based on a valid refresh token
	// 1. get request content (refresh token)
	// 2. validate refresh token through DB query
	// 3. generate access token
	// 4. return access token to user
}

func (cfg *ApiConfig) HandlerLogoutUser(w http.ResponseWriter, r *http.Request) {
	// for revoking USER refresh token
}

package api

import (
	"net/http"
)

func (cfg *ApiConfig) HandlerLoginUser(w http.ResponseWriter, r *http.Request) {
	// for authenticating USERS, not VISITORS
	// 1. get request content
	// 2. generate access token
	// 3. query for refresh token
	// 4. return access token to user
}

package api

import (
	"encoding/json"
	"net/http"
)

type VisitorRequestParameters struct {
	Name    string `json:"name"`
	Purpose string `json:"purpose"`
}

func (cfg *ApiConfig) HandlerPostVisitors(w http.ResponseWriter, r *http.Request) {
	// 1. get request data: name, purpose
	decorder := json.NewDecoder(r.Body)
	reqParams := VisitorRequestParameters{}

	// 2. make refresh token
}

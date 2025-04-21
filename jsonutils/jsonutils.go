package jsonutils

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteError(w http.ResponseWriter, respCode int, err error, msg string) {
	if err != nil {
		log.Println(err)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	e := errorResponse{Error: msg}
	WriteJSON(w, respCode, e)
}

func WriteJSON(w http.ResponseWriter, respCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(respCode)
	w.Write(dat)
}

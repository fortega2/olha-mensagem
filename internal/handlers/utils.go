package handlers

import (
	"encoding/json"
	"net/http"
)

func setContentTypeJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload any, errEncodeMsg string) {
	setContentTypeJSON(w)
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, errEncodeMsg, http.StatusInternalServerError)
	}
}

package httpUtils

import (
	"encoding/json"
	"net/http"
	"rest_api_poc/pkg/logger"
)

// WriteJSON writes JSON response
func WriteJson(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Encode directly to ResponseWriter (memory-efficient, adds newline)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// If encoding fails, send Internal Server Error
		logger.Error("Failed to encode JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// WriteStatus writes only HTTP status
func WriteStatus(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

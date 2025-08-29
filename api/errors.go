package api

import (
	"encoding/json"
	"net/http"
)

// Error represents a structured error response.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse sends a structured error response to the client.
func ErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Error{Code: code, Message: message})
}

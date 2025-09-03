package api

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Error represents a structured error response.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse sends a structured error response to the client and logs server errors.
func ErrorResponse(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Error{Code: code, Message: message})

	// Log error details for server errors
	if code >= 500 {
		logrus.WithFields(logrus.Fields{
			"error_code": code,
			"error_msg":  message,
		}).Error("Server error occurred")
	} else if code >= 400 {
		logrus.WithFields(logrus.Fields{
			"error_code": code,
			"error_msg":  message,
		}).Warn("Client error occurred")
	}
}

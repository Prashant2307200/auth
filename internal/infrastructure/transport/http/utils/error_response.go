package utils

import (
	"encoding/json"
	"net/http"
)

// ErrorCode is a typed string enum for error codes returned to clients
type ErrorCode string

const (
	BAD_REQUEST    ErrorCode = "BAD_REQUEST"
	UNAUTHORIZED   ErrorCode = "UNAUTHORIZED"
	FORBIDDEN      ErrorCode = "FORBIDDEN"
	NOT_FOUND      ErrorCode = "NOT_FOUND"
	CONFLICT       ErrorCode = "CONFLICT"
	INTERNAL_ERROR ErrorCode = "INTERNAL_ERROR"
	RATE_LIMITED   ErrorCode = "RATE_LIMITED"
)

// ErrorResponse is the standardized JSON error payload
type ErrorResponse struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// SendErrorResponse writes a standardized JSON error response
func SendErrorResponse(w http.ResponseWriter, statusCode int, code ErrorCode, message string) {
	writeJSONError(w, statusCode, &ErrorResponse{Code: code, Message: message})
}

// SendErrorResponseWithDetails writes a standardized JSON error response including details
func SendErrorResponseWithDetails(w http.ResponseWriter, statusCode int, code ErrorCode, message string, details interface{}) {
	// convert details to string (simple representation) to keep payload minimal
	var d string
	if details != nil {
		// try to marshal to JSON string if it's not already a string
		switch v := details.(type) {
		case string:
			d = v
		default:
			if b, err := json.Marshal(v); err == nil {
				d = string(b)
			}
		}
	}
	writeJSONError(w, statusCode, &ErrorResponse{Code: code, Message: message, Details: d})
}

// writeJSONError centralizes header/status and JSON encoding for both helpers
func writeJSONError(w http.ResponseWriter, statusCode int, er *ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(er)
}

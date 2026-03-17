package utils

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

// NewSuccessResponse creates a standardized success response
func NewSuccessResponse(message string, data interface{}) SuccessResponse {
	return SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates a standardized error response
func NewErrorResponse(err error) ErrorResponse {
	return ErrorResponse{
		Status: "error",
		Error:  err.Error(),
	}
}

// WriteSuccess writes a success response with status code
func WriteSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := NewSuccessResponse(message, data)
	return json.NewEncoder(w).Encode(response)
}

// WriteError writes an error response with status code
func WriteError(w http.ResponseWriter, statusCode int, err error) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := NewErrorResponse(err)
	return json.NewEncoder(w).Encode(response)
}

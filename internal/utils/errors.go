package utils

import (
	"database/sql"
	"errors"
	"fmt"
)

// Common application errors
var (
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInternalError = errors.New("internal server error")
)

// IsNotFoundError checks if error is a "not found" error
func IsNotFoundError(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || errors.Is(err, ErrNotFound)
}

// WrapError wraps an error with context
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// NewNotFoundError creates a not found error with resource name
func NewNotFoundError(resource string, identifier interface{}) error {
	return fmt.Errorf("%s with identifier %v not found: %w", resource, identifier, ErrNotFound)
}

// NewValidationError creates a validation error
func NewValidationError(field string, reason string) error {
	return fmt.Errorf("validation failed for field %s: %s: %w", field, reason, ErrInvalidInput)
}

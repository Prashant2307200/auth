package db

import (
	"context"
	"database/sql"
	"fmt"
)

// QueryRow handles common query row operations with proper error handling
func QueryRow(ctx context.Context, db *sql.DB, query string, args ...interface{}) (*sql.Row, error) {
	row := db.QueryRowContext(ctx, query, args...)
	return row, nil
}

// QueryRows handles common query operations with proper error handling
func QueryRows(ctx context.Context, db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	return rows, nil
}

// Exec handles common exec operations with proper error handling
func Exec(ctx context.Context, db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}
	return result, nil
}

// HandleNotFoundError wraps sql.ErrNoRows with a custom message
func HandleNotFoundError(err error, resource string, identifier interface{}) error {
	if err == sql.ErrNoRows {
		return fmt.Errorf("%s with identifier %v not found", resource, identifier)
	}
	return err
}

// SanitizeSearchInput sanitizes search input to prevent SQL injection
// Returns empty string if input is invalid
func SanitizeSearchInput(input string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = 100
	}
	
	// Remove potentially dangerous characters
	cleaned := ""
	for _, r := range input {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
		   (r >= '0' && r <= '9') || r == '@' || r == '.' || r == '_' || r == '-' || r == ' ' {
			cleaned += string(r)
		}
	}
	
	if len(cleaned) > maxLength {
		cleaned = cleaned[:maxLength]
	}
	
	return cleaned
}

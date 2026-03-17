package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const (
	MinUsernameLength = 3
	MaxUsernameLength = 15
	MinPasswordLength = 6
	MaxEmailLength    = 255
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if len(email) > MaxEmailLength {
		return fmt.Errorf("email must be at most %d characters", MaxEmailLength)
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("email format is invalid")
	}
	return nil
}

// ValidateUsername validates username format and length
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}
	if len(username) < MinUsernameLength {
		return fmt.Errorf("username must be at least %d characters", MinUsernameLength)
	}
	if len(username) > MaxUsernameLength {
		return fmt.Errorf("username must be at most %d characters", MaxUsernameLength)
	}
	
	// Check for valid characters (alphanumeric and underscore)
	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' {
			return fmt.Errorf("username can only contain letters, numbers, and underscores")
		}
	}
	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < MinPasswordLength {
		return fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	return nil
}

// SanitizeString removes potentially dangerous characters from input
func SanitizeString(input string, maxLength int) string {
	if maxLength <= 0 {
		maxLength = 1000
	}
	
	// Remove control characters and trim
	cleaned := strings.TrimSpace(input)
	var result strings.Builder
	result.Grow(len(cleaned))
	
	for _, r := range cleaned {
		if r >= 32 && r != 127 { // Printable ASCII except DEL
			result.WriteRune(r)
		}
	}
	
	output := result.String()
	if len(output) > maxLength {
		output = output[:maxLength]
	}
	
	return output
}

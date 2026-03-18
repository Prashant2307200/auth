package validator

import (
	"errors"
	"net/mail"
	"regexp"
)

// ValidationError represents a single field validation failure
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of ValidationError and implements error
type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	return v[0].Message
}

// ValidateEmail checks email per RFC-ish rules using net/mail parse
func ValidateEmail(email string) (bool, error) {
	if email == "" {
		return false, errors.New("email is required")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ValidatePassword enforces: min 8 chars, 1 upper, 1 lower, 1 digit, 1 special
func ValidatePassword(password string) (bool, error) {
	if len(password) < 8 {
		return false, errors.New("password must be at least 8 characters")
	}
	var (
		upper   = regexp.MustCompile(`[A-Z]`)
		lower   = regexp.MustCompile(`[a-z]`)
		digit   = regexp.MustCompile(`[0-9]`)
		special = regexp.MustCompile(`[!@#~$%^&*()_+\-={}\[\]:";'<>?,./\\|]`)
	)
	if !upper.MatchString(password) {
		return false, errors.New("password must contain at least one uppercase letter")
	}
	if !lower.MatchString(password) {
		return false, errors.New("password must contain at least one lowercase letter")
	}
	if !digit.MatchString(password) {
		return false, errors.New("password must contain at least one digit")
	}
	if !special.MatchString(password) {
		return false, errors.New("password must contain at least one special character")
	}
	return true, nil
}

// ValidateUsername ensures 3-20 chars, alphanumeric and underscore only
func ValidateUsername(username string) (bool, error) {
	if username == "" {
		return false, errors.New("username is required")
	}
	re := regexp.MustCompile(`^[A-Za-z0-9_]{3,20}$`)
	if !re.MatchString(username) {
		return false, errors.New("username must be 3-20 characters and contain only letters, numbers, or underscore")
	}
	return true, nil
}

// ValidatePhoneNumber validates E.164 format: + followed by 2-15 digits
func ValidatePhoneNumber(phone string) (bool, error) {
	if phone == "" {
		return false, errors.New("phone is required")
	}
	re := regexp.MustCompile(`^\+[1-9][0-9]{1,14}$`)
	if !re.MatchString(phone) {
		return false, errors.New("phone must be in E.164 format")
	}
	return true, nil
}

package middleware

import (
	"regexp"

	playvalidator "github.com/go-playground/validator"
)

// ValidateRequest validates a struct using go-playground/validator
// v is expected to be a pointer to a struct
func ValidateRequest(v interface{}) error {
	validate := playvalidator.New()

	// register custom 'username' validator (alphanum + underscore)
	_ = validate.RegisterValidation("username", func(fl playvalidator.FieldLevel) bool {
		re := regexp.MustCompile(`^[A-Za-z0-9_]+$`)
		return re.MatchString(fl.Field().String())
	})

	// register custom 'e164' validator for phone numbers
	_ = validate.RegisterValidation("e164", func(fl playvalidator.FieldLevel) bool {
		re := regexp.MustCompile(`^\+[1-9][0-9]{1,14}$`)
		return re.MatchString(fl.Field().String())
	})

	return validate.Struct(v)
}

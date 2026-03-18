package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Prashant2307200/auth-service/internal/utils"
	"github.com/Prashant2307200/auth-service/pkg/db"
	"github.com/go-playground/validator"
)

var validate = validator.New()

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

const (
	StatusOK    = "OK"
	StatusError = "ERROR"
)

func WriteJson(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func GeneralError(err error) Response {
	return Response{
		Status: StatusError,
		Error:  SafeClientMessage(err),
	}
}

// ErrorToStatus maps domain errors to HTTP status codes; defaults to 500.
func ErrorToStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch {
	case errors.Is(err, utils.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, utils.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, utils.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, utils.ErrNotFound), errors.Is(err, db.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// SafeClientMessage returns a generic message for clients; never leaks internal error details.
func SafeClientMessage(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, utils.ErrInvalidInput):
		return "invalid input"
	case errors.Is(err, utils.ErrUnauthorized):
		return "authentication required"
	case errors.Is(err, utils.ErrForbidden):
		return "forbidden"
	case errors.Is(err, utils.ErrNotFound), errors.Is(err, db.ErrNotFound):
		return "resource not found"
	default:
		return "internal error"
	}
}

func GeneralMessage(message string) Response {
	return Response{
		Status: StatusOK,
		Error:  message,
	}
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// WriteSuccess writes a standardized success response
func WriteSuccess(w http.ResponseWriter, statusCode int, message string, data interface{}) error {
	response := SuccessResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	}
	return WriteJson(w, statusCode, response)
}

func ValidationError[T any](entity *T) error {
	if err := validate.Struct(entity); err != nil {
		slog.Error("Validation error", slog.Any("error", err))
		if ve, ok := err.(validator.ValidationErrors); ok {
			var errs []string
			for _, e := range ve {
				switch e.ActualTag() {
				case "required":
					errs = append(errs, fmt.Sprintf("Field %s is required", e.Field()))
				case "email":
					errs = append(errs, fmt.Sprintf("Field %s must be a valid email", e.Field()))
				case "min":
					errs = append(errs, fmt.Sprintf("Field %s must be at least %s characters", e.Field(), e.Param()))
				case "max":
					errs = append(errs, fmt.Sprintf("Field %s must be at most %s characters", e.Field(), e.Param()))
				case "gte":
					errs = append(errs, fmt.Sprintf("Field %s must be greater than or equal to %s", e.Field(), e.Param()))
				case "lte":
					errs = append(errs, fmt.Sprintf("Field %s must be less than or equal to %s", e.Field(), e.Param()))
				default:
					errs = append(errs, fmt.Sprintf("Field %s is invalid", e.Field()))
				}
			}
			return errors.New("validation failed: " + strings.Join(errs, ", "))
		}
		return err
	}
	return nil
}

// FormatValidationErrors converts an error from validator into a slice of field/message maps
func FormatValidationErrors(err error) []map[string]string {
	var out []map[string]string
	if err == nil {
		return out
	}
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			var msg string
			switch e.ActualTag() {
			case "required":
				msg = "is required"
			case "email":
				msg = "must be a valid email"
			case "min":
				msg = "is too short"
			case "max":
				msg = "is too long"
			case "gte":
				msg = "must be greater than or equal to " + e.Param()
			case "lte":
				msg = "must be less than or equal to " + e.Param()
			default:
				msg = "is invalid"
			}
			out = append(out, map[string]string{"field": e.Field(), "message": msg})
		}
		return out
	}
	// Fallback: single message
	out = append(out, map[string]string{"field": "request", "message": err.Error()})
	return out
}

func SetTokenCookies(w http.ResponseWriter, accessToken, refreshToken string, env string) {

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		MaxAge:   15 * 60,
		HttpOnly: true,
		Secure:   env != "dev",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   7 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   env != "dev",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
}

func DeleteTokenCookies(w http.ResponseWriter, env string) {

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   env != "dev",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   env != "dev",
		SameSite: http.SameSiteLaxMode,
	})
}

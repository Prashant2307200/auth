package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
)

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
		Error:  err.Error(),
	}
}

func GeneralMessage(message string) Response {
	return Response{
		Status: StatusOK,
		Error:  message,
	}
}

func ValidationError[T any](entity *T) (error) {

	if err := validator.New().Struct(entity); err != nil {

		slog.Error("Validation error", slog.Any("error", err))

		if ve, ok := err.(validator.ValidationErrors); ok {
			var errs []string
			for _, err := range ve {
				switch err.ActualTag() {
				case "required":
					errs = append(errs, fmt.Sprintf("Field %s is required", err.Field()))
				case "email":
					errs = append(errs, fmt.Sprintf("Field %s must be a valid email", err.Field()))
				case "min":
					errs = append(errs, fmt.Sprintf("Field %s must be at least %s characters", err.Field(), err.Param()))
				case "max":
					errs = append(errs, fmt.Sprintf("Field %s must be at most %s characters", err.Field(), err.Param()))
				case "gte":
					errs = append(errs, fmt.Sprintf("Field %s must be greater than or equal to %s", err.Field(), err.Param()))
				case "lte":
					errs = append(errs, fmt.Sprintf("Field %s must be less than or equal to %s", err.Field(), err.Param()))
				default:
					errs = append(errs, fmt.Sprintf("Field %s is Invalid", err.Field()))
				}
			}
			return errors.New("validation failed: " + strings.Join(errs, ", "))
		} else {
			return err
		}
	}
	return nil
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
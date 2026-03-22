package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/dto"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/request"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
	"github.com/Prashant2307200/auth-service/internal/usecase"
)

type EmailVerificationHandler struct {
	UC usecase.EmailVerificationUsecase
}

func NewEmailVerificationHandler(uc usecase.EmailVerificationUsecase) *EmailVerificationHandler {
	return &EmailVerificationHandler{UC: uc}
}

func (h *EmailVerificationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /verify-email", h.verifyEmail)
	mux.HandleFunc("POST /resend-verification", h.resendVerification)
}

func (h *EmailVerificationHandler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	req, err := request.ParseJSON[dto.VerifyEmailRequest](r)
	if err != nil {
		slog.Error("Error parsing verify email request", slog.Any("error", err))
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.UC.VerifyEmail(r.Context(), req.Token)
	if err != nil {
		slog.Error("Error verifying email", slog.Any("error", err))
		switch {
		case errors.Is(err, usecase.ErrVerificationExpired):
			response.WriteError(w, http.StatusBadRequest, errors.New("verification link has expired"))
		case errors.Is(err, usecase.ErrEmailAlreadyVerified):
			response.WriteError(w, http.StatusBadRequest, errors.New("email is already verified"))
		case errors.Is(err, usecase.ErrTokenNotFound):
			response.WriteError(w, http.StatusBadRequest, errors.New("invalid verification link"))
		default:
			response.WriteError(w, http.StatusInternalServerError, errors.New("failed to verify email"))
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Email verified successfully", nil)
}

func (h *EmailVerificationHandler) resendVerification(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	err = h.UC.ResendVerification(r.Context(), userID)
	if err != nil {
		slog.Error("Error resending verification", slog.Any("error", err))
		switch {
		case errors.Is(err, usecase.ErrEmailAlreadyVerified):
			response.WriteError(w, http.StatusBadRequest, errors.New("email is already verified"))
		case errors.Is(err, usecase.ErrUserNotFound):
			response.WriteError(w, http.StatusNotFound, errors.New("user not found"))
		default:
			response.WriteError(w, http.StatusInternalServerError, errors.New("failed to send verification email"))
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Verification email sent", nil)
}

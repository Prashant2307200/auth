package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/dto"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/request"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
	"github.com/Prashant2307200/auth-service/internal/usecase"
)

type PasswordResetHandler struct {
	UC usecase.PasswordResetUsecase
}

func NewPasswordResetHandler(uc usecase.PasswordResetUsecase) *PasswordResetHandler {
	return &PasswordResetHandler{UC: uc}
}

func (h *PasswordResetHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /forgot-password", h.forgotPassword)
	mux.HandleFunc("POST /reset-password", h.resetPassword)
}

func (h *PasswordResetHandler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	req, err := request.ParseJSON[dto.ForgotPasswordRequest](r)
	if err != nil {
		slog.Error("Error parsing forgot password request", slog.Any("error", err))
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	_, err = h.UC.RequestReset(r.Context(), req.Email)
	if err != nil {
		slog.Error("Error requesting password reset", slog.Any("error", err))
	}

	response.WriteSuccess(w, http.StatusOK, "If an account with that email exists, a password reset link has been sent", nil)
}

func (h *PasswordResetHandler) resetPassword(w http.ResponseWriter, r *http.Request) {
	req, err := request.ParseJSON[dto.ResetPasswordRequest](r)
	if err != nil {
		slog.Error("Error parsing reset password request", slog.Any("error", err))
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.UC.ResetPassword(r.Context(), req.Token, req.NewPassword)
	if err != nil {
		slog.Error("Error resetting password", slog.Any("error", err))
		switch {
		case errors.Is(err, usecase.ErrTokenExpired):
			response.WriteError(w, http.StatusBadRequest, errors.New("reset link has expired"))
		case errors.Is(err, usecase.ErrTokenUsed):
			response.WriteError(w, http.StatusBadRequest, errors.New("reset link has already been used"))
		case errors.Is(err, usecase.ErrTokenNotFound):
			response.WriteError(w, http.StatusBadRequest, errors.New("invalid reset link"))
		default:
			response.WriteError(w, http.StatusInternalServerError, errors.New("failed to reset password"))
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Password has been reset successfully", nil)
}

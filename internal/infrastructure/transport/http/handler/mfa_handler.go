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
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
)

type MFAHandler struct {
	UC       usecase.MFAUsecase
	UserRepo interfaces.UserRepo
}

func NewMFAHandler(uc usecase.MFAUsecase, userRepo interfaces.UserRepo) *MFAHandler {
	return &MFAHandler{UC: uc, UserRepo: userRepo}
}

func (h *MFAHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /mfa/setup", h.setup)
	mux.HandleFunc("POST /mfa/enable", h.enable)
	mux.HandleFunc("POST /mfa/disable", h.disable)
	mux.HandleFunc("POST /mfa/verify", h.verify)
	mux.HandleFunc("POST /mfa/backup-codes", h.regenerateBackupCodes)
	mux.HandleFunc("GET /mfa/status", h.status)
}

func (h *MFAHandler) setup(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	user, err := h.UserRepo.GetById(r.Context(), userID)
	if err != nil {
		slog.Error("Error getting user for MFA setup", slog.Any("error", err))
		response.WriteError(w, http.StatusInternalServerError, errors.New("failed to get user"))
		return
	}

	result, err := h.UC.Setup(r.Context(), userID, user.Email)
	if err != nil {
		slog.Error("Error setting up MFA", slog.Any("error", err))
		switch {
		case errors.Is(err, usecase.ErrMFAAlreadyEnabled):
			response.WriteError(w, http.StatusBadRequest, errors.New("MFA is already enabled"))
		default:
			response.WriteError(w, http.StatusInternalServerError, errors.New("failed to setup MFA"))
		}
		return
	}

	response.WriteJson(w, http.StatusOK, dto.MFASetupResponse{
		Secret:    result.Secret,
		QRCodeURI: result.QRCodeURI,
	})
}

func (h *MFAHandler) enable(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	req, err := request.ParseJSON[dto.MFAVerifyRequest](r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	backupCodes, err := h.UC.Enable(r.Context(), userID, req.Code)
	if err != nil {
		slog.Error("Error enabling MFA", slog.Any("error", err))
		switch {
		case errors.Is(err, usecase.ErrMFAAlreadyEnabled):
			response.WriteError(w, http.StatusBadRequest, errors.New("MFA is already enabled"))
		case errors.Is(err, usecase.ErrInvalidTOTPCode):
			response.WriteError(w, http.StatusBadRequest, errors.New("invalid code"))
		case errors.Is(err, usecase.ErrMFASetupRequired):
			response.WriteError(w, http.StatusBadRequest, errors.New("please run setup first"))
		default:
			response.WriteError(w, http.StatusInternalServerError, errors.New("failed to enable MFA"))
		}
		return
	}

	response.WriteJson(w, http.StatusOK, map[string]interface{}{
		"message":      "MFA enabled successfully",
		"backup_codes": backupCodes,
	})
}

func (h *MFAHandler) disable(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	req, err := request.ParseJSON[dto.MFAVerifyRequest](r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.UC.Disable(r.Context(), userID, req.Code)
	if err != nil {
		slog.Error("Error disabling MFA", slog.Any("error", err))
		switch {
		case errors.Is(err, usecase.ErrMFANotEnabled):
			response.WriteError(w, http.StatusBadRequest, errors.New("MFA is not enabled"))
		case errors.Is(err, usecase.ErrInvalidTOTPCode):
			response.WriteError(w, http.StatusBadRequest, errors.New("invalid code"))
		default:
			response.WriteError(w, http.StatusInternalServerError, errors.New("failed to disable MFA"))
		}
		return
	}

	response.WriteSuccess(w, http.StatusOK, "MFA disabled successfully", nil)
}

func (h *MFAHandler) verify(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	req, err := request.ParseJSON[dto.MFAVerifyRequest](r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = h.UC.Verify(r.Context(), userID, req.Code)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidTOTPCode) {
			err = h.UC.VerifyBackupCode(r.Context(), userID, req.Code)
			if err != nil {
				response.WriteError(w, http.StatusBadRequest, errors.New("invalid code"))
				return
			}
		} else {
			slog.Error("Error verifying MFA", slog.Any("error", err))
			response.WriteError(w, http.StatusBadRequest, errors.New("invalid code"))
			return
		}
	}

	response.WriteSuccess(w, http.StatusOK, "MFA verified successfully", nil)
}

func (h *MFAHandler) regenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	req, err := request.ParseJSON[dto.MFAVerifyRequest](r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	codes, err := h.UC.RegenerateBackupCodes(r.Context(), userID, req.Code)
	if err != nil {
		slog.Error("Error regenerating backup codes", slog.Any("error", err))
		switch {
		case errors.Is(err, usecase.ErrMFANotEnabled):
			response.WriteError(w, http.StatusBadRequest, errors.New("MFA is not enabled"))
		case errors.Is(err, usecase.ErrInvalidTOTPCode):
			response.WriteError(w, http.StatusBadRequest, errors.New("invalid code"))
		default:
			response.WriteError(w, http.StatusInternalServerError, errors.New("failed to regenerate backup codes"))
		}
		return
	}

	response.WriteJson(w, http.StatusOK, map[string]interface{}{
		"message":      "Backup codes regenerated",
		"backup_codes": codes,
	})
}

func (h *MFAHandler) status(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	enabled, err := h.UC.IsEnabled(r.Context(), userID)
	if err != nil {
		slog.Error("Error checking MFA status", slog.Any("error", err))
		response.WriteError(w, http.StatusInternalServerError, errors.New("failed to check MFA status"))
		return
	}

	response.WriteJson(w, http.StatusOK, map[string]interface{}{
		"mfa_enabled": enabled,
	})
}

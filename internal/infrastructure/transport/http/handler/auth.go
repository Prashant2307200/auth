package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/dto"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	uutils "github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/request"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	v "github.com/Prashant2307200/auth-service/pkg/validator"
)

type AuthHandler struct {
	UC  *usecase.AuthUseCase
	ENV string
}

func NewAuthHandler(uc *usecase.AuthUseCase, env string) *AuthHandler {
	return &AuthHandler{UC: uc, ENV: env}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /register/", h.register)
	mux.HandleFunc("POST /login/", h.login)
	mux.HandleFunc("DELETE /logout/", h.logout)
	mux.HandleFunc("GET /profile/", h.profile)
	mux.HandleFunc("PUT /profile/", h.updateProfile)
	mux.HandleFunc("DELETE /profile/", h.deleteProfile)
	mux.HandleFunc("GET /refresh/", h.refresh)
	mux.HandleFunc("GET /public-key", h.publicKey)
	mux.HandleFunc("GET /upload-signature", h.uploadSignature)
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {

	reqDto, err := request.ParseJSON[dto.RegisterRequest](r)
	if err != nil {
		slog.Error("Error parsing JSON", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, err.Error())
		return
	}

	if err := middleware.ValidateRequest(reqDto); err != nil {
		slog.Error("Error validating JSON", slog.Any("error", err))
		// try to format into field errors
		if errs := response.FormatValidationErrors(err); len(errs) > 0 {
			response.WriteJson(w, http.StatusBadRequest, map[string]interface{}{"errors": errs})
			return
		}
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, err.Error())
		return
	}
	user := &entity.User{
		Username:   reqDto.Username,
		Email:      reqDto.Email,
		Password:   reqDto.Password,
		ProfilePic: reqDto.ProfilePic,
		Role:       reqDto.Role,
	}
	opts := &usecase.RegisterOptions{
		InviteToken:  reqDto.InviteToken,
		BusinessSlug: reqDto.BusinessSlug,
	}
	access_token, refresh_token, err := h.UC.RegisterUser(r.Context(), user, opts)
	if err != nil {
		// if validation errors, return 400 with structured errors
		var ves responseErrors
		if unwrapValidationErrors(err, &ves) {
			response.WriteJson(w, http.StatusBadRequest, map[string]interface{}{"errors": ves})
			return
		}
		slog.Error("Error registering user", slog.Any("error", err))
		status := response.ErrorToStatus(err)
		code := uutils.INTERNAL_ERROR
		if status == http.StatusBadRequest {
			code = uutils.BAD_REQUEST
		}
		uutils.SendErrorResponse(w, status, code, err.Error())
		return
	}

	response.SetTokenCookies(w, access_token, refresh_token, h.ENV)
	response.WriteSuccess(w, http.StatusOK, "user registered successfully", nil)
}

// responseErrors for sending validation errors to client
type responseErrors []struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// unwrapValidationErrors tries to extract validator.ValidationErrors from an error
func unwrapValidationErrors(err error, out *responseErrors) bool {
	var ves v.ValidationErrors
	if errors.As(err, &ves) {
		var res responseErrors
		for _, e := range ves {
			res = append(res, struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			}{Field: e.Field, Message: e.Message})
		}
		*out = res
		return true
	}
	return false
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {

	loginDto, err := request.ParseJSON[dto.LoginRequest](r)
	if err != nil {
		slog.Error("Error parsing JSON", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, err.Error())
		return
	}

	if err = middleware.ValidateRequest(loginDto); err != nil {
		slog.Error("Error validating JSON", slog.Any("error", err))
		// if validator returned field errors, format them
		if errs := response.FormatValidationErrors(err); len(errs) > 0 {
			response.WriteJson(w, http.StatusBadRequest, map[string]interface{}{"errors": errs})
			return
		}
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, err.Error())
		return
	}

	access_token, refresh_token, err := h.UC.LoginUser(r.Context(), loginDto.Email, loginDto.Password)
	if err != nil {
		// if validation errors, return 400 with structured errors
		var ves responseErrors
		if unwrapValidationErrors(err, &ves) {
			response.WriteJson(w, http.StatusBadRequest, map[string]interface{}{"errors": ves})
			return
		}
		slog.Error("Error logging in user", slog.String("email", loginDto.Email), slog.Any("error", err))
		// Don't expose whether user exists or password is wrong for security
		uutils.SendErrorResponse(w, http.StatusUnauthorized, uutils.UNAUTHORIZED, "invalid email or password")
		return
	}

	slog.Info("User logged in")
	response.SetTokenCookies(w, access_token, refresh_token, h.ENV)
	response.WriteSuccess(w, http.StatusOK, "user logged in successfully", nil)
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		slog.Error("Failed to get user ID from context", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusUnauthorized, uutils.UNAUTHORIZED, "authentication required")
		return
	}

	err = h.UC.LogoutUser(r.Context(), id)
	if err != nil {
		slog.Error("Failed to logout user", slog.Int64("user_id", id), slog.Any("error", err))
		status := response.ErrorToStatus(err)
		code := uutils.INTERNAL_ERROR
		if status == http.StatusUnauthorized {
			code = uutils.UNAUTHORIZED
		}
		uutils.SendErrorResponse(w, status, code, err.Error())
		return
	}

	response.DeleteTokenCookies(w, h.ENV)
	response.WriteSuccess(w, http.StatusOK, "user logged out successfully", nil)
}

func (h *AuthHandler) profile(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		slog.Error("Failed to get user ID from context", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusUnauthorized, uutils.UNAUTHORIZED, "authentication required")
		return
	}

	user, err := h.UC.GetAuthUserProfile(r.Context(), id)
	if err != nil {
		slog.Error("Error getting user profile", slog.Any("error", err))
		status := response.ErrorToStatus(err)
		code := uutils.INTERNAL_ERROR
		if status == http.StatusUnauthorized {
			code = uutils.UNAUTHORIZED
		}
		uutils.SendErrorResponse(w, status, code, err.Error())
		return
	}

	response.WriteJson(w, http.StatusOK, user)
}

func (h *AuthHandler) updateProfile(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		slog.Error("Failed to get user ID from context", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusUnauthorized, uutils.UNAUTHORIZED, "authentication required")
		return
	}

	req, err := request.ParseJSON[dto.ProfileUpdateRequest](r)
	if err != nil {
		slog.Error("Error parsing JSON", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, err.Error())
		return
	}

	existing, err := h.UC.GetAuthUserProfile(r.Context(), id)
	if err != nil {
		slog.Error("Error getting user profile", slog.Any("error", err))
		status := response.ErrorToStatus(err)
		code := uutils.INTERNAL_ERROR
		uutils.SendErrorResponse(w, status, code, err.Error())
		return
	}

	user := mergeProfileUpdate(existing, req)
	if err := response.ValidationError(user); err != nil {
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, err.Error())
		return
	}

	if err := h.UC.UpdateAuthUserProfile(r.Context(), id, user); err != nil {
		status := response.ErrorToStatus(err)
		code := uutils.INTERNAL_ERROR
		if status == http.StatusBadRequest {
			code = uutils.BAD_REQUEST
		}
		uutils.SendErrorResponse(w, status, code, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, "user profile updated successfully", nil)
}

func mergeProfileUpdate(existing *entity.User, req *dto.ProfileUpdateRequest) *entity.User {
	u := &entity.User{
		ID:         existing.ID,
		Username:   existing.Username,
		Email:      existing.Email,
		Password:   existing.Password,
		ProfilePic: existing.ProfilePic,
		Role:       existing.Role,
		CreatedAt:  existing.CreatedAt,
	}
	if req.Username != "" {
		u.Username = req.Username
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.ProfilePic != "" {
		u.ProfilePic = req.ProfilePic
	}
	return u
}

func (h *AuthHandler) uploadSignature(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		slog.Error("Failed to get user ID from context", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusUnauthorized, uutils.UNAUTHORIZED, "authentication required")
		return
	}

	sig, err := h.UC.GenerateUploadSignature(r.Context(), id)
	if err != nil {
		slog.Error("Error generating upload signature", slog.Any("error", err))
		status := response.ErrorToStatus(err)
		code := uutils.INTERNAL_ERROR
		uutils.SendErrorResponse(w, status, code, err.Error())
		return
	}

	response.WriteJson(w, http.StatusOK, sig)
}

func (h *AuthHandler) deleteProfile(w http.ResponseWriter, r *http.Request) {
	id, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		slog.Error("Failed to get user ID from context", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusUnauthorized, uutils.UNAUTHORIZED, "authentication required")
		return
	}

	if err := h.UC.UserRepo.DeleteById(r.Context(), id); err != nil {
		status := response.ErrorToStatus(err)
		code := uutils.INTERNAL_ERROR
		if status == http.StatusNotFound {
			code = uutils.NOT_FOUND
		}
		uutils.SendErrorResponse(w, status, code, err.Error())
		return
	}

	response.WriteSuccess(w, http.StatusOK, "user profile deleted successfully", nil)
}

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {

	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		uutils.SendErrorResponse(w, http.StatusUnauthorized, uutils.UNAUTHORIZED, "refresh token not found")
		return
	}

	refresh, access, err := h.UC.RefreshSession(r.Context(), refreshCookie.Value)
	if err != nil {
		uutils.SendErrorResponse(w, http.StatusUnauthorized, uutils.UNAUTHORIZED, err.Error())
		return
	}

	response.SetTokenCookies(w, access, refresh, h.ENV)
	response.WriteSuccess(w, http.StatusOK, "session refreshed successfully", nil)
}

func (h *AuthHandler) publicKey(w http.ResponseWriter, r *http.Request) {

	pubKey, err := h.UC.GetPublicKey()
	if err != nil {
		slog.Error("Error getting public key", slog.Any("error", err))
		status := response.ErrorToStatus(err)
		uutils.SendErrorResponse(w, status, uutils.INTERNAL_ERROR, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/x-pem-file")
	if _, err := w.Write(pubKey); err != nil {
		slog.Error("Failed to write public key response", slog.Any("error", err))
	}
}

package handler

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/request"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
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
}

func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {

	user, err := request.ParseJSON[entity.User](r)
	if err != nil {
		slog.Error("Error parsing JSON", slog.Any("error", err))
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	err = response.ValidationError(user)
	if err != nil {
		slog.Error("Error validating JSON", slog.Any("error", err))
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	access_token, refresh_token, err := h.UC.RegisterUser(r.Context(), user)
	if err != nil {
		slog.Error("Error registering user", slog.Any("error", err))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.SetTokenCookies(w, access_token, refresh_token, h.ENV)
	response.WriteJson(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "user registered successfully",
	})
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {

	user, err := request.ParseJSON[entity.Login](r)
	if err != nil {
		slog.Error("Error parsing JSON", slog.Any("error", err))
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	err = response.ValidationError(user)
	if err != nil {
		slog.Error("Error validating JSON", slog.Any("error", err))
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	access_token, refresh_token, err := h.UC.LoginUser(r.Context(), user.Email, user.Password)
	if err != nil {
		slog.Error("Error logging in user", slog.Any("error", err))
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	slog.Info("User logged in")
	response.SetTokenCookies(w, access_token, refresh_token, h.ENV)
	response.WriteJson(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "user logged in successfully",
	})
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {

	id, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		if err == sql.ErrNoRows {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(errors.New("user not found")))
		}
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	h.UC.LogoutUser(r.Context(), id)

	response.DeleteTokenCookies(w, h.ENV)

	response.WriteJson(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "user logged out successfully",
	})
}

func (h *AuthHandler) profile(w http.ResponseWriter, r *http.Request) {

	id, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		if err == sql.ErrNoRows {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(errors.New("user not found")))
		}
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	user, err := h.UC.GetAuthUserProfile(r.Context(), id)
	if err != nil {
		slog.Error("Error getting user profile", slog.Any("error", err))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.WriteJson(w, http.StatusOK, user)
}

func (h *AuthHandler) updateProfile(w http.ResponseWriter, r *http.Request) {

	id, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		if err == sql.ErrNoRows {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(errors.New("user not found")))
		}
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	user, file, fileHeader, err := request.ParseMultipartForm[entity.User](r, 10<<20, "profile_pic")
	if err != nil {
		slog.Error("Error parsing multipart form", slog.Any("error", err))
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}
	if file != nil {
		defer file.Close()
	}

	if file != nil {
		url, err := h.UC.CloudService.UploadImage(r.Context(), file, fileHeader)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		user.ProfilePic = url
	}

	if err := response.ValidationError(user); err != nil {
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	if err := h.UC.UpdateAuthUserProfile(r.Context(), id, user); err != nil {
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.WriteJson(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "user profile updated successfully",
	})
}

func (h *AuthHandler) deleteProfile(w http.ResponseWriter, r *http.Request) {

	id, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		if err == sql.ErrNoRows {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(errors.New("user not found")))
		}
		response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
		return
	}

	if err := h.UC.UserRepo.DeleteById(r.Context(), id); err != nil {
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	response.WriteJson(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "user profile deleted successfully",
	})
}

func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {

	refreshCookie, err := r.Cookie("refresh_token")
	if err != nil {
		response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(errors.New("refresh token not found")))
		return
	}

	refresh, access, err := h.UC.RefreshSession(r.Context(), refreshCookie.Value)
	if err != nil {
		response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(err))
		return
	}

	response.SetTokenCookies(w, access, refresh, h.ENV)

	response.WriteJson(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "session refreshed successfully",
	})
}

func (h *AuthHandler) publicKey(w http.ResponseWriter, r *http.Request) {

	pubKey, err := h.UC.GetPublicKey()
	if err != nil {
		slog.Error("Error getting public key", slog.Any("error", err))
		response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
		return
	}

	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Write(pubKey)
}
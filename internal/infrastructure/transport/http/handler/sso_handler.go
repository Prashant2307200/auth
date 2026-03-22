package handler

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
	"github.com/Prashant2307200/auth-service/internal/usecase"
)

type SSOHandler struct {
	UC      usecase.SSOUsecase
	ENV     string
	BaseURL string
}

func NewSSOHandler(uc usecase.SSOUsecase, env, baseURL string) *SSOHandler {
	return &SSOHandler{UC: uc, ENV: env, BaseURL: baseURL}
}

func (h *SSOHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /google", h.googleRedirect)
	mux.HandleFunc("GET /google/callback", h.googleCallback)
}

func (h *SSOHandler) googleRedirect(w http.ResponseWriter, r *http.Request) {
	state := generateStateToken()
	
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   int(10 * time.Minute / time.Second),
		HttpOnly: true,
		Secure:   h.ENV != "dev",
		SameSite: http.SameSiteLaxMode,
	})

	url := h.UC.GetGoogleAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *SSOHandler) googleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		slog.Error("Missing oauth state cookie")
		h.redirectWithError(w, r, "authentication_failed")
		return
	}

	state := r.URL.Query().Get("state")
	if state != stateCookie.Value {
		slog.Error("Invalid oauth state")
		h.redirectWithError(w, r, "authentication_failed")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		errorMsg := r.URL.Query().Get("error")
		slog.Error("OAuth error", slog.String("error", errorMsg))
		h.redirectWithError(w, r, "authentication_failed")
		return
	}

	accessToken, refreshToken, _, isNewUser, err := h.UC.HandleGoogleCallback(r.Context(), code)
	if err != nil {
		slog.Error("Error handling Google callback", slog.Any("error", err))
		switch {
		case errors.Is(err, usecase.ErrGoogleAuthFailed):
			h.redirectWithError(w, r, "authentication_failed")
		case errors.Is(err, usecase.ErrGoogleEmailMissing):
			h.redirectWithError(w, r, "email_required")
		default:
			h.redirectWithError(w, r, "server_error")
		}
		return
	}

	response.SetTokenCookies(w, accessToken, refreshToken, h.ENV)

	redirectURL := h.BaseURL + "/auth/callback"
	if isNewUser {
		redirectURL += "?new_user=true"
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (h *SSOHandler) redirectWithError(w http.ResponseWriter, r *http.Request, errorCode string) {
	redirectURL := h.BaseURL + "/auth/callback?error=" + errorCode
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func generateStateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

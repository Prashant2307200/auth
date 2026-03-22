package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
	"github.com/Prashant2307200/auth-service/internal/service"
)

type SessionHandler struct {
	SessionSvc *service.SessionService
	ENV        string
}

func NewSessionHandler(sessionSvc *service.SessionService, env string) *SessionHandler {
	return &SessionHandler{SessionSvc: sessionSvc, ENV: env}
}

func (h *SessionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /sessions", h.listSessions)
	mux.HandleFunc("DELETE /sessions/{id}", h.revokeSession)
	mux.HandleFunc("DELETE /sessions", h.revokeAllSessions)
}

func (h *SessionHandler) listSessions(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	currentSessionID := getCurrentSessionID(r)

	sessions, err := h.SessionSvc.ListUserSessions(r.Context(), userID)
	if err != nil {
		slog.Error("Error listing sessions", slog.Any("error", err))
		response.WriteError(w, http.StatusInternalServerError, errors.New("failed to list sessions"))
		return
	}

	type sessionResponse struct {
		ID         string `json:"id"`
		DeviceInfo string `json:"device_info"`
		IPAddress  string `json:"ip_address"`
		CreatedAt  string `json:"created_at"`
		LastUsedAt string `json:"last_used_at"`
		Current    bool   `json:"current"`
	}

	var resp []sessionResponse
	for _, s := range sessions {
		resp = append(resp, sessionResponse{
			ID:         s.ID,
			DeviceInfo: s.DeviceInfo,
			IPAddress:  s.IPAddress,
			CreatedAt:  s.CreatedAt.Format("2006-01-02T15:04:05Z"),
			LastUsedAt: s.LastUsedAt.Format("2006-01-02T15:04:05Z"),
			Current:    s.ID == currentSessionID,
		})
	}

	if resp == nil {
		resp = []sessionResponse{}
	}

	response.WriteJson(w, http.StatusOK, map[string]interface{}{
		"sessions": resp,
	})
}

func (h *SessionHandler) revokeSession(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		response.WriteError(w, http.StatusBadRequest, errors.New("session ID required"))
		return
	}

	currentSessionID := getCurrentSessionID(r)
	if sessionID == currentSessionID {
		response.WriteError(w, http.StatusBadRequest, errors.New("cannot revoke current session"))
		return
	}

	err = h.SessionSvc.RevokeSession(r.Context(), userID, sessionID)
	if err != nil {
		slog.Error("Error revoking session", slog.Any("error", err))
		response.WriteError(w, http.StatusInternalServerError, errors.New("failed to revoke session"))
		return
	}

	response.WriteSuccess(w, http.StatusOK, "Session revoked", nil)
}

func (h *SessionHandler) revokeAllSessions(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	currentSessionID := getCurrentSessionID(r)

	err = h.SessionSvc.RevokeAllSessions(r.Context(), userID, currentSessionID)
	if err != nil {
		slog.Error("Error revoking all sessions", slog.Any("error", err))
		response.WriteError(w, http.StatusInternalServerError, errors.New("failed to revoke sessions"))
		return
	}

	response.WriteSuccess(w, http.StatusOK, "All other sessions revoked", nil)
}

func getCurrentSessionID(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}
	return cookie.Value
}

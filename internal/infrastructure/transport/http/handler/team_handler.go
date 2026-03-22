package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
	"github.com/Prashant2307200/auth-service/internal/usecase"
)

type TeamHandler struct {
	UC  usecase.TeamUsecase
	AMW func(http.Handler) http.Handler
}

func NewTeamHandler(uc usecase.TeamUsecase, authMiddleware func(http.Handler) http.Handler) *TeamHandler {
	return &TeamHandler{UC: uc, AMW: authMiddleware}
}

// RegisterRoutes registers paths under /team/ (full URL: /api/v1/team/... after main router prefix).
func (h *TeamHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /invite", h.invite)
	mux.HandleFunc("GET /members", h.listMembers)
	mux.HandleFunc("PATCH /members/{id}/role", h.updateMemberRole)
	mux.HandleFunc("DELETE /members/{id}", h.removeMember)
	mux.HandleFunc("POST /invites/{token}/revoke", h.revokeInvitation)
}

type inviteRequest struct {
	Email string `json:"email"`
	Role  int    `json:"role"`
}

func (h *TeamHandler) invite(w http.ResponseWriter, r *http.Request) {
	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("failed to parse invite request", slog.Any("error", err))
		response.WriteError(w, http.StatusBadRequest, errors.New("bad request"))
		return
	}

	if err := h.UC.ValidateInviteEmail(req.Email); err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.UC.ValidateRole(req.Role); err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}

	businessID := middleware.GetTenantID(r)
	if businessID == 0 {
		response.WriteError(w, http.StatusUnauthorized, errors.New("tenant not found"))
		return
	}

	token, err := h.UC.InviteUser(r.Context(), businessID, req.Email, req.Role)
	if err != nil {
		slog.Error("failed to invite user", slog.Any("error", err))
		response.WriteError(w, http.StatusInternalServerError, errors.New("failed to invite user"))
		return
	}

	_ = response.WriteJson(w, http.StatusCreated, map[string]any{"invite_token": token, "message": "invitation sent"})
}

func (h *TeamHandler) listMembers(w http.ResponseWriter, r *http.Request) {
	businessID := middleware.GetTenantID(r)
	if businessID == 0 {
		response.WriteError(w, http.StatusUnauthorized, errors.New("tenant not found"))
		return
	}
	members, err := h.UC.ListMembers(r.Context(), businessID)
	if err != nil {
		slog.Error("failed to list members", slog.Any("error", err))
		response.WriteError(w, http.StatusInternalServerError, errors.New("failed to list members"))
		return
	}
	_ = response.WriteJson(w, http.StatusOK, members)
}

func (h *TeamHandler) updateMemberRole(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errors.New("invalid member id"))
		return
	}
	var body struct {
		Role int `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, errors.New("bad request"))
		return
	}
	if err := h.UC.ValidateRole(body.Role); err != nil {
		response.WriteError(w, http.StatusBadRequest, err)
		return
	}
	businessID := middleware.GetTenantID(r)
	if err := h.UC.UpdateMemberRole(r.Context(), businessID, id, body.Role); err != nil {
		slog.Error("failed to update role", slog.Any("error", err))
		response.WriteError(w, http.StatusInternalServerError, errors.New("failed to update role"))
		return
	}
	_ = response.WriteJson(w, http.StatusOK, map[string]any{"message": "updated"})
}

func (h *TeamHandler) removeMember(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errors.New("invalid member id"))
		return
	}
	businessID := middleware.GetTenantID(r)
	if err := h.UC.RemoveMember(r.Context(), businessID, id); err != nil {
		slog.Error("failed to remove member", slog.Any("error", err))
		response.WriteError(w, http.StatusInternalServerError, errors.New("failed to remove member"))
		return
	}
	_ = response.WriteJson(w, http.StatusOK, map[string]any{"message": "removed"})
}

func (h *TeamHandler) revokeInvitation(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		response.WriteError(w, http.StatusBadRequest, errors.New("token is required"))
		return
	}
	err := h.UC.RevokeInvitation(r.Context(), token)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.WriteError(w, http.StatusNotFound, errors.New("invitation not found"))
			return
		}
		if strings.Contains(err.Error(), "cannot revoke") {
			response.WriteError(w, http.StatusBadRequest, err)
			return
		}
		slog.Error("failed to revoke invitation", slog.Any("error", err))
		response.WriteError(w, http.StatusInternalServerError, errors.New("failed to revoke invitation"))
		return
	}
	_ = response.WriteJson(w, http.StatusOK, map[string]any{"message": "invitation revoked"})
}

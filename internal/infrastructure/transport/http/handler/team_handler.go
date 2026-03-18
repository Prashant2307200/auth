package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/usecase"
)

type TeamHandler struct {
	UC  usecase.TeamUsecase
	AMW func(http.Handler) http.Handler
}

func NewTeamHandler(uc usecase.TeamUsecase, authMiddleware func(http.Handler) http.Handler) *TeamHandler {
	return &TeamHandler{UC: uc, AMW: authMiddleware}
}

func (h *TeamHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/team/invite", h.invite)
	mux.HandleFunc("GET /api/v1/team/members", h.listMembers)
	mux.HandleFunc("PATCH /api/v1/team/members/", h.updateMemberRole) // expects /.../members/{id}/role
	mux.HandleFunc("DELETE /api/v1/team/members/", h.removeMember)    // expects /.../members/{id}
}

type inviteRequest struct {
	Email string `json:"email"`
	Role  int    `json:"role"`
}

func (h *TeamHandler) invite(w http.ResponseWriter, r *http.Request) {
	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("failed to parse invite request", slog.Any("error", err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	businessID := middleware.GetTenantID(r)
	if businessID == 0 {
		http.Error(w, "tenant not found", http.StatusUnauthorized)
		return
	}
	if err := h.UC.InviteUser(r.Context(), businessID, req.Email, req.Role); err != nil {
		slog.Error("failed to invite user", slog.Any("error", err))
		http.Error(w, "failed to invite user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"invited"}`))
}

func (h *TeamHandler) listMembers(w http.ResponseWriter, r *http.Request) {
	businessID := middleware.GetTenantID(r)
	if businessID == 0 {
		http.Error(w, "tenant not found", http.StatusUnauthorized)
		return
	}
	members, err := h.UC.ListMembers(r.Context(), businessID)
	if err != nil {
		slog.Error("failed to list members", slog.Any("error", err))
		http.Error(w, "failed to list members", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(members)
}

func (h *TeamHandler) updateMemberRole(w http.ResponseWriter, r *http.Request) {
	// path like /api/v1/team/members/{id}/role
	parts := splitPath(r.URL.Path)
	if len(parts) < 6 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[5]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid member id", http.StatusBadRequest)
		return
	}
	var body struct {
		Role int `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	businessID := middleware.GetTenantID(r)
	if err := h.UC.UpdateMemberRole(r.Context(), businessID, id, body.Role); err != nil {
		slog.Error("failed to update role", slog.Any("error", err))
		http.Error(w, "failed to update role", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"updated"}`))
}

func (h *TeamHandler) removeMember(w http.ResponseWriter, r *http.Request) {
	// path like /api/v1/team/members/{id}
	parts := splitPath(r.URL.Path)
	if len(parts) < 5 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	idStr := parts[4]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid member id", http.StatusBadRequest)
		return
	}
	businessID := middleware.GetTenantID(r)
	if err := h.UC.RemoveMember(r.Context(), businessID, id); err != nil {
		slog.Error("failed to remove member", slog.Any("error", err))
		http.Error(w, "failed to remove member", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"removed"}`))
}

func splitPath(p string) []string {
	// simple split preserving empties
	var res []string
	cur := ""
	for i := 0; i < len(p); i++ {
		if p[i] == '/' {
			res = append(res, cur)
			cur = ""
			continue
		}
		cur += string(p[i])
	}
	res = append(res, cur)
	return res
}

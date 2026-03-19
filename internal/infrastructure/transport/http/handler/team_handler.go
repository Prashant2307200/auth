package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	uutils "github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils"
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
	mux.HandleFunc("POST /api/v1/team/invites/", h.revokeInvitation)  // expects /.../invites/{token}/revoke
}

type inviteRequest struct {
	Email string `json:"email"`
	Role  int    `json:"role"`
}

func (h *TeamHandler) invite(w http.ResponseWriter, r *http.Request) {
	var req inviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("failed to parse invite request", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, "bad request")
		return
	}

	if err := h.UC.ValidateInviteEmail(req.Email); err != nil {
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, err.Error())
		return
	}

	if err := h.UC.ValidateRole(req.Role); err != nil {
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, err.Error())
		return
	}

	businessID := middleware.GetTenantID(r)
	if businessID == 0 {
		uutils.SendErrorResponse(w, http.StatusUnauthorized, uutils.UNAUTHORIZED, "tenant not found")
		return
	}

	token, err := h.UC.InviteUser(r.Context(), businessID, req.Email, req.Role)
	if err != nil {
		slog.Error("failed to invite user", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusInternalServerError, uutils.INTERNAL_ERROR, "failed to invite user")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{"invite_token": token, "message": "invitation sent"})
}

func (h *TeamHandler) listMembers(w http.ResponseWriter, r *http.Request) {
	businessID := middleware.GetTenantID(r)
	if businessID == 0 {
		uutils.SendErrorResponse(w, http.StatusUnauthorized, uutils.UNAUTHORIZED, "tenant not found")
		return
	}
	members, err := h.UC.ListMembers(r.Context(), businessID)
	if err != nil {
		slog.Error("failed to list members", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusInternalServerError, uutils.INTERNAL_ERROR, "failed to list members")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(members)
}

func (h *TeamHandler) updateMemberRole(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(r.URL.Path)
	if len(parts) < 6 {
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, "invalid path")
		return
	}
	idStr := parts[5]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, "invalid member id")
		return
	}
	var body struct {
		Role int `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, "bad request")
		return
	}
	if err := h.UC.ValidateRole(body.Role); err != nil {
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, err.Error())
		return
	}
	businessID := middleware.GetTenantID(r)
	if err := h.UC.UpdateMemberRole(r.Context(), businessID, id, body.Role); err != nil {
		slog.Error("failed to update role", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusInternalServerError, uutils.INTERNAL_ERROR, "failed to update role")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"updated"}`))
}

func (h *TeamHandler) removeMember(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(r.URL.Path)
	if len(parts) < 5 {
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, "invalid path")
		return
	}
	idStr := parts[4]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, "invalid member id")
		return
	}
	businessID := middleware.GetTenantID(r)
	if err := h.UC.RemoveMember(r.Context(), businessID, id); err != nil {
		slog.Error("failed to remove member", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusInternalServerError, uutils.INTERNAL_ERROR, "failed to remove member")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"removed"}`))
}

func (h *TeamHandler) revokeInvitation(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 7 || parts[6] != "revoke" {
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, "invalid request")
		return
	}
	token := parts[5]
	if token == "" {
		uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, "token is required")
		return
	}
	err := h.UC.RevokeInvitation(r.Context(), token)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			uutils.SendErrorResponse(w, http.StatusNotFound, uutils.NOT_FOUND, "invitation not found")
			return
		}
		if strings.Contains(err.Error(), "cannot revoke") {
			uutils.SendErrorResponse(w, http.StatusBadRequest, uutils.BAD_REQUEST, err.Error())
			return
		}
		slog.Error("failed to revoke invitation", slog.Any("error", err))
		uutils.SendErrorResponse(w, http.StatusInternalServerError, uutils.INTERNAL_ERROR, "failed to revoke invitation")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "invitation revoked"})
}

func splitPath(p string) []string {
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

// local email validation helper
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	at := strings.Index(email, "@")
	if at <= 0 || at >= len(email)-3 {
		return false
	}
	dot := strings.LastIndex(email, ".")
	if dot <= at+1 || dot >= len(email)-1 {
		return false
	}
	return true
}

package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/repository"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
)

type AuditHandler struct {
	AuditRepo repository.AuditRepository
}

func NewAuditHandler(auditRepo repository.AuditRepository) *AuditHandler {
	return &AuditHandler{AuditRepo: auditRepo}
}

func (h *AuditHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /audit-logs", h.listAuditLogs)
}

func (h *AuditHandler) listAuditLogs(w http.ResponseWriter, r *http.Request) {
	_, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}

	businessID, err := middleware.GetBusinessIDFromContext(r.Context())
	if err != nil || businessID == 0 {
		response.WriteError(w, http.StatusBadRequest, errors.New("business context required"))
		return
	}

	query := r.URL.Query()

	page := 1
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := (page - 1) * limit

	action := query.Get("action")
	fromTime := query.Get("from")
	toTime := query.Get("to")

	var userID *int64
	if userIDStr := query.Get("user_id"); userIDStr != "" {
		if parsed, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			userID = &parsed
		}
	}

	logs, err := h.AuditRepo.ListWithFilter(r.Context(), businessID, userID, action, fromTime, toTime, limit, offset)
	if err != nil {
		slog.Error("Error listing audit logs", slog.Any("error", err))
		response.WriteError(w, http.StatusInternalServerError, errors.New("failed to list audit logs"))
		return
	}

	if logs == nil {
		logs = []*entity.AuditLog{}
	}

	response.WriteJson(w, http.StatusOK, map[string]interface{}{
		"audit_logs": logs,
		"page":       page,
		"limit":      limit,
	})
}

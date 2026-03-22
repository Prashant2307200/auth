package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Prashant2307200/auth-service/internal/usecase"
)

type HealthHandler struct {
	UC *usecase.HealthUseCase
}

func NewHealthHandler(uc *usecase.HealthUseCase) *HealthHandler {
	return &HealthHandler{UC: uc}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	hs, _ := h.UC.Check(ctx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// ensure timestamp is RFC3339
	type out struct {
		Status    string `json:"status"`
		Database  string `json:"database"`
		Redis     string `json:"redis"`
		Timestamp string `json:"timestamp"`
	}
	o := out{
		Status:    hs.Status,
		Database:  hs.Database,
		Redis:     hs.Redis,
		Timestamp: hs.Timestamp.Format(time.RFC3339),
	}
	_ = json.NewEncoder(w).Encode(o)
}

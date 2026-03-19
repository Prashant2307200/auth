package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// SystemHealthHandler performs lightweight health checks for live/ready endpoints
type SystemHealthHandler struct {
	db  *sql.DB
	rdb *redis.Client
}

// NewSystemHealthHandler constructs the handler and accepts concrete DB and Redis clients
func NewSystemHealthHandler(db *sql.DB, rdb *redis.Client) *SystemHealthHandler {
	return &SystemHealthHandler{db: db, rdb: rdb}
}

// Live returns basic liveness status (process + DB reachable)
func (h *SystemHealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	status := http.StatusOK
	if h.db != nil {
		if err := h.db.PingContext(ctx); err != nil {
			status = http.StatusServiceUnavailable
		}
	}

	w.WriteHeader(status)
	w.Write([]byte("ok"))
}

// Ready returns readiness information including DB and Redis statuses
func (h *SystemHealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	type serviceStatus struct {
		Database string `json:"database"`
		Redis    string `json:"redis"`
	}
	type out struct {
		Status    string        `json:"status"`
		Timestamp string        `json:"timestamp"`
		Services  serviceStatus `json:"services"`
	}

	st := out{Timestamp: time.Now().UTC().Format(time.RFC3339)}
	st.Status = "healthy"
	st.Services.Database = "ok"
	st.Services.Redis = "ok"

	// DB check
	if h.db == nil {
		st.Status = "degraded"
		st.Services.Database = "down"
	} else {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()
		if err := h.db.PingContext(ctx); err != nil {
			st.Status = "degraded"
			st.Services.Database = "down"
		}
	}

	// Redis check
	if h.rdb == nil {
		st.Status = "degraded"
		st.Services.Redis = "down"
	} else {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()
		if err := h.rdb.Ping(ctx).Err(); err != nil {
			st.Status = "degraded"
			st.Services.Redis = "down"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(st)
}

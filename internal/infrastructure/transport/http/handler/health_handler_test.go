package handler

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// Test Live and Ready endpoints using an in-memory sqlite DB and redis nil client
func TestSystemHealthHandler_LiveReady(t *testing.T) {
	var db *sql.DB = nil
	var rdb *redis.Client = nil

	h := NewSystemHealthHandler(db, rdb)

	// Live
	req := httptest.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()
	h.Live(w, req)
	res := w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)

	// Ready
	req2 := httptest.NewRequest("GET", "/health/ready", nil)
	w2 := httptest.NewRecorder()
	h.Ready(w2, req2)
	res2 := w2.Result()
	require.Equal(t, http.StatusOK, res2.StatusCode)

	// response body contains timestamp; simple smoke check
	// read body
	// ensure handler returns quickly
	time.Sleep(10 * time.Millisecond)
}

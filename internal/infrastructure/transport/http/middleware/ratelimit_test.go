package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimit_UnderLimit(t *testing.T) {
	handler := RateLimit(5, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/login/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "request %d should pass", i+1)
	}
}

func TestRateLimit_ExceedsLimit(t *testing.T) {
	handler := RateLimit(3, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/login/", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	}

	// 4th request should be rate limited
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/login/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("Retry-After"))
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	handler := RateLimit(1, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First IP
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/login/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Second IP should not be affected
	rr2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/login/", nil)
	req2.RemoteAddr = "5.6.7.8:5678"
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)
}

func TestRateLimit_WindowReset(t *testing.T) {
	handler := RateLimit(1, 50*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/login/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Exceed limit
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)

	// Wait for window to reset
	time.Sleep(60 * time.Millisecond)

	rr3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/login/", nil)
	req3.RemoteAddr = "1.2.3.4:1234"
	handler.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusOK, rr3.Code)
}

package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	limiter := NewRateLimiter(1, 1)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:8080"

	tests := []struct {
		name     string
		allowed  bool
		waitTime time.Duration
	}{
		{"first request allowed", true, 0},
		{"second request denied", false, 0},
		{"after refill allowed", true, 1100 * time.Millisecond},
	}

	for _, tt := range tests {
		if tt.waitTime > 0 {
			time.Sleep(tt.waitTime)
		}

		if result := limiter.Allow(req); result != tt.allowed {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.allowed, result)
		}
	}
}

func TestRateLimiter_GetIP(t *testing.T) {
	limiter := NewRateLimiter(1, 1)

	tests := []struct {
		name       string
		xForwarded string
		remoteAddr string
		expectedIP string
	}{
		{"X-Forwarded-For present", "203.0.113.1", "192.168.1.1:8080", "203.0.113.1"},
		{"X-Forwarded-For empty", "", "192.168.1.1:8080", "192.168.1.1"},
		{"no port in remoteAddr", "", "192.168.1.1", "192.168.1.1"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/", nil)
		if tt.xForwarded != "" {
			req.Header.Set("X-Forwarded-For", tt.xForwarded)
		}
		req.RemoteAddr = tt.remoteAddr

		if ip := limiter.getIP(req); ip != tt.expectedIP {
			t.Errorf("%s: expected %s, got %s", tt.name, tt.expectedIP, ip)
		}
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	limiter := NewRateLimiter(10, 1)

	req1 := httptest.NewRequest("GET", "/", nil)
	req1.RemoteAddr = "192.168.1.1:8080"

	req2 := httptest.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "192.168.1.2:8080"

	limiter.Allow(req1)
	limiter.Allow(req2)

	limiter.mu.RLock()
	if len(limiter.limiters) != 2 {
		t.Errorf("expected 2 limiters, got %d", len(limiter.limiters))
	}
	limiter.mu.RUnlock()

	limiter.Cleanup(100 * time.Millisecond)
	limiter.mu.RLock()
	if len(limiter.limiters) != 2 {
		t.Errorf("cleanup with max age 100ms should keep recent limiters, got %d", len(limiter.limiters))
	}
	limiter.mu.RUnlock()

	time.Sleep(150 * time.Millisecond)
	limiter.Cleanup(100 * time.Millisecond)
	limiter.mu.RLock()
	if len(limiter.limiters) != 0 {
		t.Errorf("cleanup with max age 100ms should remove old limiters, got %d", len(limiter.limiters))
	}
	limiter.mu.RUnlock()
}

func TestRateLimiter_Middleware(t *testing.T) {
	limiter := NewRateLimiter(1, 2)

	handler := limiter.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const testIP = "192.168.1.1:8080"

	tests := []struct {
		name           string
		expectedStatus int
	}{
		{"first request", http.StatusOK},
		{"second request", http.StatusOK},
		{"third request exceeds limit", http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = testIP
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != tt.expectedStatus {
			t.Errorf("%s: expected status %d, got %d", tt.name, tt.expectedStatus, w.Code)
		}
	}
}

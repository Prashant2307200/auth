package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prashant2307200/auth-service/pkg/ratelimit"
)

func TestWrapRateLimitedRoutes_PrefixMatch_Allows(t *testing.T) {
	rl := ratelimit.NewRateLimiter(1000, 10)
	handler := wrapRateLimitedRoutes(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), rl, []string{"/register/"})

	req := httptest.NewRequest("GET", "/register/foo", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected allowed request to return 200, got %d", w.Code)
	}
}

func TestWrapRateLimitedRoutes_PrefixMatch_DeniesWhenExceeded(t *testing.T) {
	// very restrictive limiter that should deny immediate subsequent requests
	rl := ratelimit.NewRateLimiter(0, 0)
	handler := wrapRateLimitedRoutes(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), rl, []string{"/register/"})

	req := httptest.NewRequest("GET", "/register/", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	// First request may be allowed depending on rate limiter semantics; send a second one to observe denial.
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code == http.StatusOK {
		t.Fatalf("expected second request to be rate-limited, got %d", w2.Code)
	}
}

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

func TestWrapRateLimitedRoutes_Table(t *testing.T) {
	cases := []struct {
		name       string
		rateLimit  *ratelimit.RateLimiter
		routes     []string
		url        string
		remoteAddr string
		expectOK   bool
	}{
		{
			name:      "allow_on_prefix_match",
			rateLimit: ratelimit.NewRateLimiter(1000, 10),
			routes:    []string{"/register/", "/login/"},
			url:       "/register/foo",
			expectOK:  true,
		},
		{
			name:       "deny_when_exceeded_exact",
			rateLimit:  ratelimit.NewRateLimiter(0, 0),
			routes:     []string{"/register/"},
			url:        "/register/",
			remoteAddr: "2.2.2.2:1234",
			expectOK:   false,
		},
		{
			name:      "non_matching_path_passes_through",
			rateLimit: ratelimit.NewRateLimiter(0, 0),
			routes:    []string{"/register/"},
			url:       "/health",
			expectOK:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			handler := wrapRateLimitedRoutes(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}), tc.rateLimit, tc.routes)

			req := httptest.NewRequest("GET", tc.url, nil)
			if tc.remoteAddr != "" {
				req.RemoteAddr = tc.remoteAddr
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if tc.expectOK && w.Code != http.StatusOK {
				t.Fatalf("expected OK for %s, got %d", tc.name, w.Code)
			}
			if !tc.expectOK && w.Code == http.StatusOK {
				t.Fatalf("expected rate-limited for %s, got %d", tc.name, w.Code)
			}
		})
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "DENY", rr.Header().Get("X-Frame-Options"))
	require.Equal(t, "strict-origin-when-cross-origin", rr.Header().Get("Referrer-Policy"))
}

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestID_GeneratesID(t *testing.T) {
	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	handler.ServeHTTP(rr, req)

	assert.NotEmpty(t, capturedID)
	assert.Equal(t, capturedID, rr.Header().Get("X-Request-ID"))
}

func TestRequestID_UsesExistingHeader(t *testing.T) {
	var capturedID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")
	handler.ServeHTTP(rr, req)

	assert.Equal(t, "my-custom-id", capturedID)
	assert.Equal(t, "my-custom-id", rr.Header().Get("X-Request-ID"))
}

func TestGetRequestID_EmptyContext(t *testing.T) {
	id := GetRequestID(context.Background())
	assert.Empty(t, id)
}

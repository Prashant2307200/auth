package logging

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		inputRequestID  string
		expectGenerated bool
	}{
		{"with provided request ID", "custom-id-123", false},
		{"without request ID (should generate)", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.inputRequestID != "" {
				req.Header.Set("X-Request-ID", tt.inputRequestID)
			}

			rr := httptest.NewRecorder()
			handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rid := GetRequestID(r.Context())
				if tt.inputRequestID != "" {
					assert.Equal(t, tt.inputRequestID, rid)
				} else {
					assert.NotEmpty(t, rid)
				}
			}))

			handler.ServeHTTP(rr, req)

			responseID := rr.Header().Get("X-Request-ID")
			assert.NotEmpty(t, responseID)
			if tt.inputRequestID != "" {
				assert.Equal(t, tt.inputRequestID, responseID)
			}
		})
	}
}

func TestWithUserID(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, 123)

	userID := GetUserID(ctx)
	assert.Equal(t, int64(123), userID)
}

func TestGetRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), RequestIDKey, "test-id-456")
	rid := GetRequestID(ctx)
	assert.Equal(t, "test-id-456", rid)
}

func TestGetRequestIDEmpty(t *testing.T) {
	ctx := context.Background()
	rid := GetRequestID(ctx)
	assert.Empty(t, rid)
}

func TestGetUserIDEmpty(t *testing.T) {
	ctx := context.Background()
	uid := GetUserID(ctx)
	assert.Equal(t, int64(0), uid)
}

func TestGetUserIDWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "not-an-int64")
	uid := GetUserID(ctx)
	assert.Equal(t, int64(0), uid)
}

func TestRequestIDGenerationIsUnique(t *testing.T) {
	req1 := httptest.NewRequest("GET", "/test", nil)
	req2 := httptest.NewRequest("GET", "/test", nil)

	var id1, id2 string

	handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r.Context())
		if id1 == "" {
			id1 = id
		} else {
			id2 = id
		}
	}))

	handler.ServeHTTP(httptest.NewRecorder(), req1)
	handler.ServeHTTP(httptest.NewRecorder(), req2)

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestContextPropagation(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, RequestIDKey, "req-123")
	ctx = WithUserID(ctx, 999)

	assert.Equal(t, "req-123", GetRequestID(ctx))
	assert.Equal(t, int64(999), GetUserID(ctx))
}

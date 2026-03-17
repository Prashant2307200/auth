package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Prashant2307200/auth-service/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestRSAKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	// Generate a test RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey, &privateKey.PublicKey
}

func createTestTokenService(t *testing.T) *service.JWTTokenService {
	privateKey, publicKey := generateTestRSAKeys(t)
	// Use a mock Redis client - in real tests you'd use a test container or mock
	// For now, we'll use nil and only test token verification which doesn't need Redis
	return &service.JWTTokenService{
		PublicAccessSecret: publicKey,
		AccessSecret:       privateKey,
		RefreshSecret:      "test-refresh-secret",
		Rdb:                nil, // Not needed for token verification tests
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() context.Context
		wantErr  bool
		wantID   int64
	}{
		{
			name: "valid user ID in context",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), userContextKey, int64(123))
			},
			wantErr: false,
			wantID:  123,
		},
		{
			name: "no user ID in context",
			setupCtx: func() context.Context {
				return context.Background()
			},
			wantErr: true,
		},
		{
			name: "wrong type in context",
			setupCtx: func() context.Context {
				return context.WithValue(context.Background(), userContextKey, "not-an-int64")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx()
			userID, err := GetUserIDFromContext(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Zero(t, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, userID)
			}
		})
	}
}

func TestAuthenticate(t *testing.T) {
	tokenService := createTestTokenService(t)

	tests := []struct {
		name           string
		path           string
		setupRequest   func() *http.Request
		wantStatusCode int
		wantUserID     int64
	}{
		{
			name: "public login route",
			path: "/api/v1/auth/login/",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/api/v1/auth/login/", nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "public register route",
			path: "/api/v1/auth/register/",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/api/v1/auth/register/", nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "public refresh route",
			path: "/api/v1/auth/refresh/",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/api/v1/auth/refresh/", nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "public key route",
			path: "/api/v1/auth/public-key",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/api/v1/auth/public-key", nil)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "protected route without token",
			path: "/api/v1/auth/profile/",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/api/v1/auth/profile/", nil)
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "protected route with invalid token",
			path: "/api/v1/auth/profile/",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/api/v1/auth/profile/", nil)
				req.AddCookie(&http.Cookie{Name: "access_token", Value: "invalid-token"})
				return req
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "protected route with valid token",
			path: "/api/v1/auth/profile/",
			setupRequest: func() *http.Request {
				// Create a valid token
				claims := jwt.MapClaims{
					"userId": float64(123),
					"exp":    time.Now().Add(15 * time.Minute).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				tokenString, err := token.SignedString(tokenService.AccessSecret)
				require.NoError(t, err)

				req := httptest.NewRequest("GET", "/api/v1/auth/profile/", nil)
				req.AddCookie(&http.Cookie{Name: "access_token", Value: tokenString})
				return req
			},
			wantStatusCode: http.StatusOK,
			wantUserID:     123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()
			rr := httptest.NewRecorder()

			handler := Authenticate(tokenService, "test")
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.wantUserID > 0 {
					userID, err := GetUserIDFromContext(r.Context())
					assert.NoError(t, err)
					assert.Equal(t, tt.wantUserID, userID)
				}
				w.WriteHeader(http.StatusOK)
			})

			handler(nextHandler).ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatusCode, rr.Code)
		})
	}
}

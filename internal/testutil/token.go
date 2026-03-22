package testutil

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/service"
	"github.com/stretchr/testify/require"
)

// NewTestTokenService creates a JWTTokenService with ephemeral RSA keys for testing.
func NewTestTokenService(t *testing.T) *service.JWTTokenService {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return &service.JWTTokenService{
		PublicAccessSecret: &privateKey.PublicKey,
		AccessSecret:       privateKey,
		RefreshSecret:      "test-refresh-secret",
		Rdb:                nil,
	}
}

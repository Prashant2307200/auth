package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	authgrpc "github.com/Prashant2307200/auth-service/internal/transport/grpc/proto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestGRPC_E2E_TokenVerificationWorkflow verifies end-to-end token verification flow
// Simulates: HTTP layer generates token → gRPC service verifies it
func TestGRPC_E2E_TokenVerificationWorkflow(t *testing.T) {
	// Setup: Create test dependencies (as would be initialized in cmd/main)
	jwtService := testutil.NewTestTokenService(t)
	userRepo := &testutil.MockUserRepo{}
	user := &entity.User{ID: 1, TenantID: 0, Role: 0}
	userRepo.On("GetById", mock.Anything, int64(1)).Return(user, nil)

	// Create gRPC token service instance
	grpcTokenSvc := NewTokenService(jwtService, userRepo)

	// Scenario 1: HTTP layer generates token (e.g., after login)
	token, err := jwtService.GenerateAccessToken(1)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Scenario 2: Another service calls gRPC VerifyToken
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := grpcTokenSvc.VerifyToken(ctx, &authgrpc.VerifyTokenRequest{Token: token})
	require.NoError(t, err)
	require.Equal(t, int64(1), resp.UserId)
	userRepo.AssertExpectations(t)
}

// TestGRPC_E2E_PublicKeyDistributionWorkflow verifies public key distribution for downstream verification
func TestGRPC_E2E_PublicKeyDistributionWorkflow(t *testing.T) {
	businessRepo := &testutil.MockBusinessRepo{}
	business := &entity.Business{ID: 1, PublicKey: "test-pem-key"}
	businessRepo.On("GetById", mock.Anything, int64(1)).Return(business, nil)

	// Create gRPC public key service
	grpcKeySvc := NewPublicKeyService(businessRepo)

	// Scenario: Service requests public key to verify tokens independently
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := grpcKeySvc.GetPublicKey(ctx, &authgrpc.GetPublicKeyRequest{BusinessId: 1})
	require.NoError(t, err)
	require.Equal(t, "test-pem-key", resp.PemKey)
	businessRepo.AssertExpectations(t)
}

// TestGRPC_E2E_TokenVerificationWithDifferentUsers verifies isolation between users
func TestGRPC_E2E_TokenVerificationWithDifferentUsers(t *testing.T) {
	jwtService := testutil.NewTestTokenService(t)
	userRepo := &testutil.MockUserRepo{}

	// User 1 token
	user1 := &entity.User{ID: 1, TenantID: 0, Role: 0}
	userRepo.On("GetById", mock.Anything, int64(1)).Return(user1, nil)
	token1, err := jwtService.GenerateAccessToken(1)
	require.NoError(t, err)

	// User 2 token
	user2 := &entity.User{ID: 2, TenantID: 0, Role: 0}
	userRepo.On("GetById", mock.Anything, int64(2)).Return(user2, nil)
	token2, err := jwtService.GenerateAccessToken(2)
	require.NoError(t, err)

	grpcTokenSvc := NewTokenService(jwtService, userRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Verify user 1 token
	resp1, err := grpcTokenSvc.VerifyToken(ctx, &authgrpc.VerifyTokenRequest{Token: token1})
	require.NoError(t, err)
	require.Equal(t, int64(1), resp1.UserId)

	// Verify user 2 token
	resp2, err := grpcTokenSvc.VerifyToken(ctx, &authgrpc.VerifyTokenRequest{Token: token2})
	require.NoError(t, err)
	require.Equal(t, int64(2), resp2.UserId)

	userRepo.AssertExpectations(t)
}

// TestGRPC_E2E_InvalidTokenRejection verifies gRPC properly rejects invalid tokens
func TestGRPC_E2E_InvalidTokenRejection(t *testing.T) {
	jwtService := testutil.NewTestTokenService(t)
	userRepo := &testutil.MockUserRepo{}

	grpcTokenSvc := NewTokenService(jwtService, userRepo)

	tests := []struct {
		name  string
		token string
	}{
		{"malformed", "not.a.token"},
		{"empty", ""},
		{"junk", "random-junk-string"},
		{"incomplete", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_, err := grpcTokenSvc.VerifyToken(ctx, &authgrpc.VerifyTokenRequest{Token: tt.token})
			require.Error(t, err, fmt.Sprintf("expected error for %s token", tt.name))
		})
	}
}

// TestGRPC_E2E_PublicKeyCaching verifies public key caching reduces DB calls
func TestGRPC_E2E_PublicKeyCaching(t *testing.T) {
	businessRepo := &testutil.MockBusinessRepo{}
	business := &entity.Business{ID: 1, PublicKey: "cached-pem"}
	// Setup: only called once, but GetPublicKey called twice (caching should prevent 2nd call)
	businessRepo.On("GetById", mock.Anything, int64(1)).Return(business, nil).Once()

	grpcKeySvc := NewPublicKeyService(businessRepo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First call hits DB
	resp1, err := grpcKeySvc.GetPublicKey(ctx, &authgrpc.GetPublicKeyRequest{BusinessId: 1})
	require.NoError(t, err)
	require.Equal(t, "cached-pem", resp1.PemKey)

	// Second call should use cache, not call DB again
	resp2, err := grpcKeySvc.GetPublicKey(ctx, &authgrpc.GetPublicKeyRequest{BusinessId: 1})
	require.NoError(t, err)
	require.Equal(t, "cached-pem", resp2.PemKey)

	// Verify GetById was called exactly once (caching worked)
	businessRepo.AssertExpectations(t)
}

// TestGRPC_E2E_TenantIsolation verifies gRPC respects tenant boundaries
func TestGRPC_E2E_TenantIsolation(t *testing.T) {
	jwtService := testutil.NewTestTokenService(t)
	userRepo := &testutil.MockUserRepo{}

	// User belongs to tenant 1
	user := &entity.User{ID: 1, TenantID: 1, Role: 0}
	userRepo.On("GetById", mock.Anything, int64(1)).Return(user, nil)

	grpcTokenSvc := NewTokenService(jwtService, userRepo)
	token, _ := jwtService.GenerateAccessToken(1)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to access with different tenant ID - should fail
	_, err := grpcTokenSvc.VerifyToken(ctx, &authgrpc.VerifyTokenRequest{
		Token:    token,
		TenantId: 2, // different from user's TenantID (1)
	})
	require.Error(t, err, "should reject cross-tenant access")
}

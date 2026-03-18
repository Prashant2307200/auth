package server

import (
	"context"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	authgrpc "github.com/Prashant2307200/auth-service/internal/transport/grpc/proto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestVerifyToken_Success(t *testing.T) {
	jwt := testutil.NewTestTokenService(t)
	userRepo := &testutil.MockUserRepo{}
	user := &entity.User{ID: 1, TenantID: 0, Role: 0}
	userRepo.On("GetById", mock.Anything, int64(1)).Return(user, nil)
	svc := NewTokenService(jwt, userRepo)

	token, err := jwt.GenerateAccessToken(1)
	require.NoError(t, err)

	res, err := svc.VerifyToken(context.Background(), &authgrpc.VerifyTokenRequest{Token: token})
	require.NoError(t, err)
	require.Equal(t, int64(1), res.UserId)
}

func TestVerifyToken_InvalidToken(t *testing.T) {
	jwt := testutil.NewTestTokenService(t)
	userRepo := &testutil.MockUserRepo{}
	svc := NewTokenService(jwt, userRepo)

	_, err := svc.VerifyToken(context.Background(), &authgrpc.VerifyTokenRequest{Token: "bad-token"})
	require.Error(t, err)
}

func TestVerifyToken_TenantMismatch(t *testing.T) {
	jwt := testutil.NewTestTokenService(t)
	userRepo := &testutil.MockUserRepo{}
	user := &entity.User{ID: 1, TenantID: 99, Role: 0}
	userRepo.On("GetById", mock.Anything, int64(1)).Return(user, nil)
	svc := NewTokenService(jwt, userRepo)

	token, err := jwt.GenerateAccessToken(1)
	require.NoError(t, err)

	_, err = svc.VerifyToken(context.Background(), &authgrpc.VerifyTokenRequest{Token: token, TenantId: 1})
	require.Error(t, err)
}

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

func TestGetPublicKey_SuccessAndCaching(t *testing.T) {
	busRepo := &testutil.MockBusinessRepo{}
	bus := &entity.Business{ID: 1, PublicKey: "pem-data"}
	busRepo.On("GetById", mock.Anything, int64(1)).Return(bus, nil).Once()
	svc := NewPublicKeyService(busRepo)

	res, err := svc.GetPublicKey(context.Background(), &authgrpc.GetPublicKeyRequest{BusinessId: 1})
	require.NoError(t, err)
	require.Equal(t, "pem-data", res.PemKey)

	// update repo to return different key but cache should keep old
	busRepo.On("GetById", mock.Anything, int64(1)).Return(&entity.Business{ID: 1, PublicKey: "pem-new"}, nil).Once()
	res2, err := svc.GetPublicKey(context.Background(), &authgrpc.GetPublicKeyRequest{BusinessId: 1})
	require.NoError(t, err)
	require.Equal(t, "pem-data", res2.PemKey)

	// expire cache
	svc.mu.Lock()
	v := svc.cache[1]
	v.createdAt = time.Now().Add(-10 * time.Minute)
	svc.cache[1] = v
	svc.mu.Unlock()

	res3, err := svc.GetPublicKey(context.Background(), &authgrpc.GetPublicKeyRequest{BusinessId: 1})
	require.NoError(t, err)
	require.Equal(t, "pem-new", res3.PemKey)
}

func TestGetPublicKey_NotFound(t *testing.T) {
	busRepo := &testutil.MockBusinessRepo{}
	busRepo.On("GetById", mock.Anything, int64(2)).Return(nil, fmt.Errorf("not found"))
	svc := NewPublicKeyService(busRepo)

	_, err := svc.GetPublicKey(context.Background(), &authgrpc.GetPublicKeyRequest{BusinessId: 2})
	require.Error(t, err)
}

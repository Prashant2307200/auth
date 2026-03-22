package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	authgrpc "github.com/Prashant2307200/auth-service/internal/transport/grpc/proto"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
)

type cachedKey struct {
	key       string
	createdAt time.Time
}

type PublicKeyService struct {
	authgrpc.UnimplementedPublicKeyServiceServer
	businessRepo interfaces.BusinessRepo
	mu           sync.RWMutex
	cache        map[int64]cachedKey
	ttl          time.Duration
}

func NewPublicKeyService(businessRepo interfaces.BusinessRepo) *PublicKeyService {
	return &PublicKeyService{businessRepo: businessRepo, cache: make(map[int64]cachedKey), ttl: 5 * time.Minute}
}

func (s *PublicKeyService) GetPublicKey(ctx context.Context, req *authgrpc.GetPublicKeyRequest) (*authgrpc.GetPublicKeyResponse, error) {
	bid := req.GetBusinessId()
	s.mu.RLock()
	if v, ok := s.cache[bid]; ok {
		if time.Since(v.createdAt) < s.ttl {
			s.mu.RUnlock()
			return &authgrpc.GetPublicKeyResponse{PemKey: v.key, CreatedAt: v.createdAt.Format(time.RFC3339)}, nil
		}
	}
	s.mu.RUnlock()

	b, err := s.businessRepo.GetById(ctx, bid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch business: %w", err)
	}

	pem := b.PublicKey

	s.mu.Lock()
	s.cache[bid] = cachedKey{key: pem, createdAt: time.Now()}
	s.mu.Unlock()

	return &authgrpc.GetPublicKeyResponse{PemKey: pem, CreatedAt: time.Now().Format(time.RFC3339)}, nil
}

package server

import (
	"context"
	"fmt"

	"github.com/Prashant2307200/auth-service/internal/entity"
	authgrpc "github.com/Prashant2307200/auth-service/internal/transport/grpc/proto"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
)

type TokenService struct {
	jwtService interfaces.TokenService
	userRepo   interfaces.UserRepo
}

func NewTokenService(jwtService interfaces.TokenService, userRepo interfaces.UserRepo) *TokenService {
	return &TokenService{jwtService: jwtService, userRepo: userRepo}
}

func roleNameFromUser(u *entity.User) string {
	if u.RoleName != "" {
		return u.RoleName
	}
	if u.Role == entity.RoleAdmin {
		return "admin"
	}
	return "user"
}

func (s *TokenService) VerifyToken(ctx context.Context, req *authgrpc.VerifyTokenRequest) (*authgrpc.VerifyTokenResponse, error) {
	userID, err := s.jwtService.VerifyToken(ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("token verification failed: %w", err)
	}

	user, err := s.userRepo.GetById(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if req.TenantId != 0 && user.TenantID != req.TenantId {
		return nil, fmt.Errorf("tenant mismatch: token tenant %d != requested %d", user.TenantID, req.TenantId)
	}

	return &authgrpc.VerifyTokenResponse{UserId: userID, Role: roleNameFromUser(user)}, nil
}

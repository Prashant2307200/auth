package handler

import (
	"context"
	"os"

	authpb "github.com/Prashant2307200/auth-service/internal/infrastructure/transport/rpc/proto"
)

type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
}

func (s *AuthServer) GetPublicKey(ctx context.Context, _ *authpb.Empty) (*authpb.PublicKeyResponse, error) {
	keyBytes, err := os.ReadFile("keys/public.pem")
	if err != nil {
		return nil, err
	}

	return &authpb.PublicKeyResponse{Pem: string(keyBytes)}, nil
}

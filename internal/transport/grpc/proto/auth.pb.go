package authgrpc

import (
	"context"
)

// VerifyTokenRequest is request for verifying token
type VerifyTokenRequest struct {
	Token    string
	TenantId int64
}

// VerifyTokenResponse is response for VerifyToken
type VerifyTokenResponse struct {
	UserId int64
	Role   string
}

// GetPublicKeyRequest placeholder to keep package cohesive (from other proto)
type GetPublicKeyRequest struct {
	BusinessId int64
}

// GetPublicKeyResponse placeholder
type GetPublicKeyResponse struct {
	PemKey    string
	CreatedAt string
}

// TokenServiceServer is the server API for TokenService service.
type TokenServiceServer interface {
	VerifyToken(ctx context.Context, req *VerifyTokenRequest) (*VerifyTokenResponse, error)
}

// PublicKeyServiceServer is the server API for PublicKeyService service.
type PublicKeyServiceServer interface {
	GetPublicKey(ctx context.Context, req *GetPublicKeyRequest) (*GetPublicKeyResponse, error)
}

package service

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

type JWTTokenService struct {
	PublicAccessSecret *rsa.PublicKey
	AccessSecret       *rsa.PrivateKey
	RefreshSecret      string
	Rdb                *redis.Client
}

func NewJWTTokenService(rdb *redis.Client, publicAccessSecretPath, accessSecretPath, refreshSecret string) (*JWTTokenService, error) {
	accessSecretBytes, err := os.ReadFile(accessSecretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file %s: %w", accessSecretPath, err)
	}

	accessSecret, err := jwt.ParseRSAPrivateKeyFromPEM(accessSecretBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key from %s: %w", accessSecretPath, err)
	}

	publicAccessSecretBytes, err := os.ReadFile(publicAccessSecretPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file %s: %w", publicAccessSecretPath, err)
	}

	publicAccessSecret, err := jwt.ParseRSAPublicKeyFromPEM(publicAccessSecretBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key from %s: %w", publicAccessSecretPath, err)
	}

	return &JWTTokenService{
		PublicAccessSecret: publicAccessSecret,
		AccessSecret:       accessSecret,
		RefreshSecret:      refreshSecret,
		Rdb:                rdb,
	}, nil
}

func (s *JWTTokenService) GenerateRefreshToken(userID int64) (string, error) {
	return generateJWT(fmt.Sprint(userID), s.RefreshSecret, 7*24*time.Hour)
}

func (s *JWTTokenService) StoreRefreshToken(ctx context.Context, userID int64, token string) error {
	return s.Rdb.Set(ctx, fmt.Sprint(userID), token, 7*24*time.Hour).Err()
}

func (s *JWTTokenService) GetRefreshToken(ctx context.Context, userID int64) (string, error) {
	return s.Rdb.Get(ctx, fmt.Sprint(userID)).Result()
}

func (s *JWTTokenService) RemoveRefreshToken(ctx context.Context, userID int64) error {
	return s.Rdb.Del(ctx, fmt.Sprint(userID)).Err()
}

func generateJWT(userID string, secret string, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(ttl).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func (s *JWTTokenService) VerifyRefreshToken(ctx context.Context, tokenStr string) (string, error) {

	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.RefreshSecret), nil
	})
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return "", fmt.Errorf("token invalid: %w", err)
	}

	userID, ok := claims["userId"].(string)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	return userID, err
}

func (s *JWTTokenService) GenerateAccessToken(userID int64, businessID ...int64) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(15 * time.Minute).Unix(),
	}
	if len(businessID) > 0 {
		claims["businessId"] = businessID[0]
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.AccessSecret)
}

func (s *JWTTokenService) VerifyToken(ctx context.Context, tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.PublicAccessSecret, nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return 0, errors.New("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("failed to extract claims from token")
	}

	userIDFloat, ok := claims["userId"].(float64)
	if !ok {
		return 0, fmt.Errorf("userId claim not found or invalid type in token")
	}

	return int64(userIDFloat), nil
}

func (s *JWTTokenService) GetPublicKeyPEM() ([]byte, error) {
	data, err := os.ReadFile("keys/public.pem")
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}
	return data, nil
}

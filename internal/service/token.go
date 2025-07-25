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
	AccessSecret  *rsa.PrivateKey
	RefreshSecret string
	Rdb           *redis.Client
}

func NewJWTTokenService(rdb *redis.Client, publicAccessSecretPath, accessSecretPath, refreshSecret string) *JWTTokenService {

	accessSecretBytes, err := os.ReadFile(accessSecretPath)
	if err != nil {
		panic(err)
	}
	accessSecret, err := jwt.ParseRSAPrivateKeyFromPEM(accessSecretBytes)
	if err != nil {
		panic(err)
	}

	publicAccessSecretBytes, err := os.ReadFile(publicAccessSecretPath)
	if err != nil {
		panic(err)
	}
	publicAccessSecret, err := jwt.ParseRSAPublicKeyFromPEM(publicAccessSecretBytes)
	if err != nil {
		panic(err)
	}

	return &JWTTokenService{
		PublicAccessSecret: publicAccessSecret,
		AccessSecret:  accessSecret,
		RefreshSecret: refreshSecret,
		Rdb:           rdb,
	}
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

// func (s *JWTTokenService) GenerateAccessToken(userID int64) (string, error) {
// 	return generateJWT(fmt.Sprint(userID), s.AccessSecret, 15*time.Minute)
// }

// func (s *JWTTokenService) VerifyAccessToken(ctx context.Context, tokenStr string) (string, error) {
	
// 	claims := jwt.MapClaims{}

// 	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
// 		return []byte(s.AccessSecret), nil
// 	})
// 	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
	// 		return "", fmt.Errorf("token invalid: %w", err)
	// 	}

// 	userID, ok := claims["userId"].(string)
// 	if !ok {
// 		return "", errors.New("invalid token claims")
// 	}

// 	return userID, err
// }

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


func (s *JWTTokenService) GenerateAccessToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(15 * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.AccessSecret)
}

func (s *JWTTokenService) VerifyToken(ctx context.Context, tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return s.PublicAccessSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid claims")
	}

	userIDFloat, ok := claims["userId"].(float64)
	if !ok {
		return 0, errors.New("userId not found")
	}

	return int64(userIDFloat), nil
}

func (s *JWTTokenService) GetPublicKeyPEM() ([]byte, error) {
	return os.ReadFile("keys/public.pem")
}
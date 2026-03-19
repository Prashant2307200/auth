package service

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
)

type JWTTokenService struct {
	PublicAccessSecret *rsa.PublicKey
	AccessSecret       *rsa.PrivateKey
	RefreshSecret      string
	Rdb                *redis.Client
	publicKeyPEM       []byte
	metrics            *TokenMetrics
}

type TokenMetrics struct {
	VerificationsTotal   prometheus.Counter
	VerificationDuration prometheus.Histogram
}

func NewTokenMetrics() (*TokenMetrics, error) {
	verifTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_token_verifications_total",
		Help: "Total number of token verification attempts",
	})

	verifDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "auth_token_verification_duration_seconds",
		Help:    "Token verification duration in seconds",
		Buckets: prometheus.DefBuckets,
	})

	return &TokenMetrics{
		VerificationsTotal:   verifTotal,
		VerificationDuration: verifDuration,
	}, nil
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

	metrics, _ := NewTokenMetrics()

	return &JWTTokenService{
		PublicAccessSecret: publicAccessSecret,
		AccessSecret:       accessSecret,
		RefreshSecret:      refreshSecret,
		Rdb:                rdb,
		publicKeyPEM:       publicAccessSecretBytes,
		metrics:            metrics,
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

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.RefreshSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", jwt.ErrTokenExpired
		}
		return "", fmt.Errorf("token invalid: %w", err)
	}
	if !token.Valid {
		return "", errors.New("token is not valid")
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
	tracer := otel.Tracer("auth-service")
	ctx, span := tracer.Start(ctx, "token.VerifyToken")
	defer span.End()

	start := time.Now()
	defer func() {
		if s.metrics != nil {
			s.metrics.VerificationDuration.Observe(time.Since(start).Seconds())
		}
	}()

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.PublicAccessSecret, nil
	})
	if err != nil {
		if s.metrics != nil {
			s.metrics.VerificationsTotal.Inc()
		}
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		if s.metrics != nil {
			s.metrics.VerificationsTotal.Inc()
		}
		return 0, errors.New("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		if s.metrics != nil {
			s.metrics.VerificationsTotal.Inc()
		}
		return 0, errors.New("failed to extract claims from token")
	}

	userIDFloat, ok := claims["userId"].(float64)
	if !ok {
		if s.metrics != nil {
			s.metrics.VerificationsTotal.Inc()
		}
		return 0, fmt.Errorf("userId claim not found or invalid type in token")
	}

	if s.metrics != nil {
		s.metrics.VerificationsTotal.Inc()
	}

	return int64(userIDFloat), nil
}

func (s *JWTTokenService) GetPublicKeyPEM() ([]byte, error) {
	return s.publicKeyPEM, nil
}

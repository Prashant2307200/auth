package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
	"time"

	// fmt was removed; keep imports minimal
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

func setupKeys(t *testing.T) (string, string) {
	t.Helper()
	// generate RSA keys and write PEM files
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privPem := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	dir := t.TempDir()
	privF := dir + "/priv.pem"
	pubF := dir + "/pub.pem"
	if err := os.WriteFile(privF, privPem, 0600); err != nil {
		t.Fatalf("failed to write priv key: %v", err)
	}
	if err := os.WriteFile(pubF, pubPem, 0600); err != nil {
		t.Fatalf("failed to write pub key: %v", err)
	}
	return pubF, privF
}

func TestGenerateAndVerifyRefreshToken(t *testing.T) {
	pub, priv := setupKeys(t)
	// NewJWTTokenService requires an *redis.Client but these tests don't use it
	var rdb *redis.Client = nil
	s, err := NewJWTTokenService(rdb, pub, priv, "refresh-secret-test")
	if err != nil {
		t.Fatalf("failed to create token service: %v", err)
	}

	token, err := s.GenerateRefreshToken(123)
	if err != nil {
		t.Fatalf("generate refresh token failed: %v", err)
	}

	// Verify valid token
	uid, err := s.VerifyRefreshToken(context.Background(), token)
	if err != nil {
		t.Fatalf("verify refresh token failed: %v", err)
	}
	if uid != "123" {
		t.Fatalf("expected uid 123 got %s", uid)
	}

	// expired token
	short, _ := generateJWT("1", "refresh-secret-test", -time.Hour)
	_, err = s.VerifyRefreshToken(context.Background(), short)
	if err == nil {
		t.Fatalf("expected error for expired token")
	}

	// invalid signature
	bad, _ := generateJWT("1", "wrong-secret", time.Hour)
	_, err = s.VerifyRefreshToken(context.Background(), bad)
	if err == nil {
		t.Fatalf("expected error for invalid signature")
	}
}

func TestGenerateAccessAndVerifyToken(t *testing.T) {
	pub, priv := setupKeys(t)
	var rdb2 *redis.Client = nil
	s, err := NewJWTTokenService(rdb2, pub, priv, "refresh-secret-test")
	if err != nil {
		t.Fatalf("failed to create token service: %v", err)
	}

	token, err := s.GenerateAccessToken(55)
	if err != nil {
		t.Fatalf("generate access token failed: %v", err)
	}

	// Tamper token: change a char
	if len(token) < 10 {
		t.Fatalf("token too short")
	}
	// verify works
	_, err = s.VerifyToken(context.Background(), token)
	if err != nil {
		t.Fatalf("verify token failed: %v", err)
	}

	// use invalid signing method token
	tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"userId": 1})
	signed, _ := tkn.SignedString([]byte("abc"))
	_, err = s.VerifyToken(context.Background(), signed)
	if err == nil {
		t.Fatalf("expected error for wrong signing method")
	}
}

func TestAccessToken_TableDriven(t *testing.T) {
	pub, priv := setupKeys(t)
	var rdb *redis.Client = nil
	s, err := NewJWTTokenService(rdb, pub, priv, "refresh-secret-test")
	if err != nil {
		t.Fatalf("failed to create token service: %v", err)
	}

	cases := []struct {
		name      string
		makeToken func() (string, error)
		wantErr   bool
		wantUser  int64
		wantBiz   *int64
	}{
		{
			name: "valid no business",
			makeToken: func() (string, error) {
				return s.GenerateAccessToken(55)
			},
			wantErr:  false,
			wantUser: 55,
			wantBiz:  nil,
		},
		{
			name: "valid with business",
			makeToken: func() (string, error) {
				return s.GenerateAccessToken(55, 99)
			},
			wantErr:  false,
			wantUser: 55,
			wantBiz:  func() *int64 { v := int64(99); return &v }(),
		},
		{
			name: "expired token",
			makeToken: func() (string, error) {
				claims := jwt.MapClaims{"userId": 77, "exp": time.Now().Add(-time.Hour).Unix()}
				tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				return tok.SignedString(s.AccessSecret)
			},
			wantErr: true,
		},
		{
			name: "wrong signing algo",
			makeToken: func() (string, error) {
				tkn := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"userId": 1})
				return tkn.SignedString([]byte("abc"))
			},
			wantErr: true,
		},
		{
			name: "missing userId",
			makeToken: func() (string, error) {
				claims := jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}
				tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				return tok.SignedString(s.AccessSecret)
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tok, err := tc.makeToken()
			if err != nil {
				t.Fatalf("makeToken failed: %v", err)
			}
			uid, err := s.VerifyToken(context.Background(), tok)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none for case %s", tc.name)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected verify error: %v", err)
			}
			if uid != tc.wantUser {
				t.Fatalf("expected user %d got %d", tc.wantUser, uid)
			}

			// If business claim expected, parse token to check claim exists
			if tc.wantBiz != nil {
				parsed, err := jwt.Parse(tok, func(token *jwt.Token) (interface{}, error) {
					return s.PublicAccessSecret, nil
				})
				if err != nil || !parsed.Valid {
					t.Fatalf("failed to parse token for business claim: %v", err)
				}
				claims := parsed.Claims.(jwt.MapClaims)
				biz, ok := claims["businessId"].(float64)
				if !ok {
					t.Fatalf("businessId claim missing or wrong type: %v", claims["businessId"])
				}
				if int64(biz) != *tc.wantBiz {
					t.Fatalf("expected business %d got %d", *tc.wantBiz, int64(biz))
				}
			}
		})
	}
}

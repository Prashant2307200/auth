package invitetoken

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type InviteTokenClaims struct {
	MemberID   int64  `json:"member_id"`
	BusinessID int64  `json:"business_id"`
	Email      string `json:"email"`
	jwt.RegisteredClaims
}

type Generator struct {
	secret         string
	expiryDuration time.Duration
}

func NewGenerator(secret string, expiryHours int) *Generator {
	return &Generator{secret: secret, expiryDuration: time.Duration(expiryHours) * time.Hour}
}

func (g *Generator) Generate(memberID, businessID int64, email string) (string, time.Time, error) {
	expiresAt := time.Now().Add(g.expiryDuration)
	claims := InviteTokenClaims{
		MemberID:   memberID,
		BusinessID: businessID,
		Email:      email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(g.secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, expiresAt, nil
}

func (g *Generator) Validate(tokenString string) (*InviteTokenClaims, error) {
	claims := &InviteTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(g.secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}
	return claims, nil
}

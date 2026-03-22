package invitetoken

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	gen := NewGenerator("test-secret", 24)
	token, expiry, err := gen.Generate(1, 10, "user@example.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.False(t, expiry.IsZero())
	assert.True(t, expiry.After(time.Now()))
}

func TestValidate_Success(t *testing.T) {
	gen := NewGenerator("test-secret", 24)
	token, _, _ := gen.Generate(1, 10, "user@example.com")
	claims, err := gen.Validate(token)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), claims.MemberID)
	assert.Equal(t, int64(10), claims.BusinessID)
	assert.Equal(t, "user@example.com", claims.Email)
}

func TestValidate_InvalidToken(t *testing.T) {
	gen := NewGenerator("test-secret", 24)
	_, err := gen.Validate("invalid-token")
	assert.Error(t, err)
}

func TestValidate_ExpiredToken(t *testing.T) {
	gen := NewGenerator("test-secret", -1)
	token, _, _ := gen.Generate(1, 10, "user@example.com")
	_, err := gen.Validate(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestValidate_WrongSecret(t *testing.T) {
	gen1 := NewGenerator("secret1", 24)
	gen2 := NewGenerator("secret2", 24)
	token, _, _ := gen1.Generate(1, 10, "user@example.com")
	_, err := gen2.Validate(token)
	assert.Error(t, err)
}

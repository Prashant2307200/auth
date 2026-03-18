package middleware

import (
	"testing"

	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/dto"
)

func TestValidateRequest_Valid(t *testing.T) {
	r := &dto.RegisterRequest{
		Username: "user_one",
		Email:    "test@example.com",
		Password: "strongPass1",
	}
	if err := ValidateRequest(r); err != nil {
		t.Fatalf("expected valid, got err=%v", err)
	}
}

func TestValidateRequest_Invalid(t *testing.T) {
	r := &dto.RegisterRequest{
		Username: "u",
		Email:    "bad-email",
		Password: "short",
	}
	if err := ValidateRequest(r); err == nil {
		t.Fatalf("expected validation error")
	}
}

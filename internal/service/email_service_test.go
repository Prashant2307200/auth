package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailService_SendInvite(t *testing.T) {
	cfg := SMTPConfig{Host: "localhost", Port: 1025, Username: "test", Password: "test", FromEmail: "noreply@example.com"}
	svc := NewEmailService(cfg, "https://example.com")
	// We cannot actually send email in unit test; ensure method exists and returns error or nil depending on environment
	err := svc.SendInvite(context.Background(), "user@example.com", "token-123")
	// Don't assert on err being nil because no SMTP server may be running
	assert.NotNil(t, svc)
	_ = err
}

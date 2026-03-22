package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const mailerooAPIURL = "https://smtp.maileroo.com/api/v2/emails"

type MailerooConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
	BaseURL   string
}

type MailerooService struct {
	cfg    MailerooConfig
	client *http.Client
}

func NewMailerooService(cfg MailerooConfig) *MailerooService {
	return &MailerooService{
		cfg: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type emailAddress struct {
	Address     string `json:"address"`
	DisplayName string `json:"display_name,omitempty"`
}

type mailerooRequest struct {
	From    emailAddress `json:"from"`
	To      emailAddress `json:"to"`
	Subject string       `json:"subject"`
	HTML    string       `json:"html"`
	Plain   string       `json:"plain,omitempty"`
}

func (m *MailerooService) send(ctx context.Context, to, subject, html, plain string) error {
	if m.cfg.APIKey == "" {
		return nil
	}

	req := mailerooRequest{
		From: emailAddress{
			Address:     m.cfg.FromEmail,
			DisplayName: m.cfg.FromName,
		},
		To: emailAddress{
			Address: to,
		},
		Subject: subject,
		HTML:    html,
		Plain:   plain,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal email request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, mailerooAPIURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Api-Key", m.cfg.APIKey)

	resp, err := m.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("send email request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("maileroo returned status %d", resp.StatusCode)
	}

	return nil
}

func (m *MailerooService) SendInvite(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("%s/accept-invite?token=%s", m.cfg.BaseURL, token)
	subject := "You've been invited to join"
	html := fmt.Sprintf(`
		<h1>Team Invitation</h1>
		<p>You've been invited to join a team. Click the link below to accept:</p>
		<p><a href="%s">Accept Invitation</a></p>
		<p>This link will expire in 24 hours.</p>
	`, link)
	plain := fmt.Sprintf("You've been invited to join a team. Accept here: %s", link)

	return m.send(ctx, to, subject, html, plain)
}

func (m *MailerooService) SendPasswordReset(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("%s/reset-password?token=%s", m.cfg.BaseURL, token)
	subject := "Reset Your Password"
	html := fmt.Sprintf(`
		<h1>Password Reset</h1>
		<p>You requested to reset your password. Click the link below:</p>
		<p><a href="%s">Reset Password</a></p>
		<p>This link will expire in 1 hour. If you didn't request this, ignore this email.</p>
	`, link)
	plain := fmt.Sprintf("Reset your password here: %s (expires in 1 hour)", link)

	return m.send(ctx, to, subject, html, plain)
}

func (m *MailerooService) SendEmailVerification(ctx context.Context, to, token string) error {
	link := fmt.Sprintf("%s/verify-email?token=%s", m.cfg.BaseURL, token)
	subject := "Verify Your Email"
	html := fmt.Sprintf(`
		<h1>Email Verification</h1>
		<p>Please verify your email address by clicking the link below:</p>
		<p><a href="%s">Verify Email</a></p>
		<p>This link will expire in 24 hours.</p>
	`, link)
	plain := fmt.Sprintf("Verify your email here: %s", link)

	return m.send(ctx, to, subject, html, plain)
}

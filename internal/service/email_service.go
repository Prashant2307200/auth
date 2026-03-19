package service

import (
	"context"
	"fmt"
	"net/smtp"
)

type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
}

type EmailService struct {
	config  SMTPConfig
	baseURL string
}

func NewEmailService(cfg SMTPConfig, baseURL string) *EmailService {
	return &EmailService{config: cfg, baseURL: baseURL}
}

func (s *EmailService) SendInvite(ctx context.Context, toEmail string, inviteToken string) error {
	inviteLink := fmt.Sprintf("%s/accept-invite?token=%s", s.baseURL, inviteToken)

	subject := "You're invited to join our workspace"
	body := fmt.Sprintf(`<html><body><h2>You're invited!</h2><p>Click <a href="%s">here</a> to accept the invitation.</p></body></html>`, inviteLink)

	return s.send(toEmail, subject, body)
}

func (s *EmailService) send(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", s.config.FromEmail, to, subject, body)
	if err := smtp.SendMail(addr, auth, s.config.FromEmail, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

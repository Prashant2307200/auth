package service

import "context"

// NoopEmailService implements usecase.EmailService without sending mail (dev / tests).
type NoopEmailService struct{}

func (NoopEmailService) SendInvite(_ context.Context, _, _ string) error {
	return nil
}

func (NoopEmailService) SendPasswordReset(_ context.Context, _, _ string) error {
	return nil
}

func (NoopEmailService) SendEmailVerification(_ context.Context, _, _ string) error {
	return nil
}

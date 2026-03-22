package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/repository"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
)

var (
	ErrEmailAlreadyVerified = errors.New("email is already verified")
	ErrVerificationExpired  = errors.New("verification link has expired")
)

type EmailVerificationUsecase interface {
	SendVerification(ctx context.Context, userID int64, email string) error
	VerifyEmail(ctx context.Context, token string) error
	ResendVerification(ctx context.Context, userID int64) error
}

type emailVerificationUsecase struct {
	userRepo     interfaces.UserRepo
	verifyRepo   repository.EmailVerificationRepository
	emailService EmailService
	auditRepo    repository.AuditRepository
}

func NewEmailVerificationUsecase(
	userRepo interfaces.UserRepo,
	verifyRepo repository.EmailVerificationRepository,
	emailService EmailService,
	auditRepo ...repository.AuditRepository,
) EmailVerificationUsecase {
	uc := &emailVerificationUsecase{
		userRepo:     userRepo,
		verifyRepo:   verifyRepo,
		emailService: emailService,
	}
	if len(auditRepo) > 0 {
		uc.auditRepo = auditRepo[0]
	}
	return uc
}

func (u *emailVerificationUsecase) SendVerification(ctx context.Context, userID int64, email string) error {
	rawToken, err := generateSecureToken(32)
	if err != nil {
		return err
	}

	tokenHash := hashVerificationToken(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)

	_ = u.verifyRepo.DeleteAllForUser(ctx, userID)

	_, err = u.verifyRepo.Create(ctx, userID, tokenHash, expiresAt)
	if err != nil {
		return err
	}

	if u.emailService != nil {
		return u.emailService.SendEmailVerification(ctx, email, rawToken)
	}

	return nil
}

func (u *emailVerificationUsecase) VerifyEmail(ctx context.Context, token string) error {
	tokenHash := hashVerificationToken(token)

	verifyToken, err := u.verifyRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return ErrTokenNotFound
	}

	if verifyToken.IsExpired() {
		return ErrVerificationExpired
	}

	user, err := u.userRepo.GetById(ctx, verifyToken.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	if err := u.userRepo.MarkEmailVerified(ctx, verifyToken.UserID); err != nil {
		return err
	}

	_ = u.verifyRepo.DeleteAllForUser(ctx, verifyToken.UserID)

	u.logAudit(ctx, verifyToken.UserID, entity.AuditActionUserEmailVerified)

	return nil
}

func (u *emailVerificationUsecase) ResendVerification(ctx context.Context, userID int64) error {
	user, err := u.userRepo.GetById(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	return u.SendVerification(ctx, userID, user.Email)
}

func (u *emailVerificationUsecase) logAudit(ctx context.Context, userID int64, action string) {
	if u.auditRepo == nil {
		return
	}
	if err := u.auditRepo.Log(ctx, &entity.AuditLog{
		UserID: userID,
		Action: action,
	}); err != nil {
		slog.Error("failed to log audit event", slog.String("action", action), slog.Int64("user_id", userID), slog.Any("error", err))
	}
}

func hashVerificationToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

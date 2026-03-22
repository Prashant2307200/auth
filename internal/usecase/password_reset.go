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
	"github.com/Prashant2307200/auth-service/pkg/hash"
)

var (
	ErrTokenExpired  = errors.New("token has expired")
	ErrTokenUsed     = errors.New("token has already been used")
	ErrTokenNotFound = errors.New("token not found")
	ErrUserNotFound  = errors.New("user not found")
)

type PasswordResetUsecase interface {
	RequestReset(ctx context.Context, email string) (token string, err error)
	ResetPassword(ctx context.Context, token string, newPassword string) error
}

type passwordResetUsecase struct {
	userRepo     interfaces.UserRepo
	resetRepo    repository.PasswordResetRepository
	emailService EmailService
	tokenService interfaces.TokenService
	auditRepo    repository.AuditRepository
}

func NewPasswordResetUsecase(
	userRepo interfaces.UserRepo,
	resetRepo repository.PasswordResetRepository,
	emailService EmailService,
	tokenService interfaces.TokenService,
	auditRepo ...repository.AuditRepository,
) PasswordResetUsecase {
	uc := &passwordResetUsecase{
		userRepo:     userRepo,
		resetRepo:    resetRepo,
		emailService: emailService,
		tokenService: tokenService,
	}
	if len(auditRepo) > 0 {
		uc.auditRepo = auditRepo[0]
	}
	return uc
}

func (u *passwordResetUsecase) RequestReset(ctx context.Context, email string) (string, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil
	}
	if user == nil {
		return "", nil
	}

	rawToken, err := generateSecureToken(32)
	if err != nil {
		return "", err
	}

	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err = u.resetRepo.Create(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return "", err
	}

	if u.emailService != nil {
		_ = u.emailService.SendPasswordReset(ctx, email, rawToken)
	}

	u.logAudit(ctx, user.ID, entity.AuditActionUserPasswordResetRequested)

	return rawToken, nil
}

func (u *passwordResetUsecase) ResetPassword(ctx context.Context, token string, newPassword string) error {
	tokenHash := hashToken(token)

	resetToken, err := u.resetRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return ErrTokenNotFound
	}

	if resetToken.IsExpired() {
		return ErrTokenExpired
	}

	if resetToken.IsUsed() {
		return ErrTokenUsed
	}

	hashedPassword, err := hash.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := u.userRepo.UpdatePassword(ctx, resetToken.UserID, hashedPassword); err != nil {
		return err
	}

	if err := u.resetRepo.MarkUsed(ctx, resetToken.ID); err != nil {
		return err
	}

	_ = u.tokenService.RemoveRefreshToken(ctx, resetToken.UserID)

	u.logAudit(ctx, resetToken.UserID, entity.AuditActionUserPasswordResetCompleted)

	return nil
}

func (u *passwordResetUsecase) logAudit(ctx context.Context, userID int64, action string) {
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

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

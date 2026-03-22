package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/repository"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
	"github.com/Prashant2307200/auth-service/pkg/encrypt"
)

var (
	ErrMFANotEnabled     = errors.New("MFA is not enabled for this user")
	ErrMFAAlreadyEnabled = errors.New("MFA is already enabled for this user")
	ErrInvalidTOTPCode   = errors.New("invalid TOTP code")
	ErrInvalidBackupCode = errors.New("invalid backup code")
	ErrMFASetupRequired  = errors.New("MFA setup required")
	ErrMFARequired       = errors.New("MFA verification required")
)

const (
	TOTPIssuer       = "AuthService"
	BackupCodeCount  = 10
	BackupCodeLength = 8
)

type MFASetupResult struct {
	Secret      string   `json:"secret"`
	QRCodeURI   string   `json:"qr_code_uri"`
	BackupCodes []string `json:"backup_codes"`
}

type MFAUsecase interface {
	Setup(ctx context.Context, userID int64, email string) (*MFASetupResult, error)
	Enable(ctx context.Context, userID int64, code string) ([]string, error)
	Disable(ctx context.Context, userID int64, code string) error
	Verify(ctx context.Context, userID int64, code string) error
	VerifyBackupCode(ctx context.Context, userID int64, code string) error
	RegenerateBackupCodes(ctx context.Context, userID int64, code string) ([]string, error)
	IsEnabled(ctx context.Context, userID int64) (bool, error)
}

type mfaUsecase struct {
	userRepo      interfaces.UserRepo
	mfaRepo       repository.MFARepository
	auditRepo     repository.AuditRepository
	encryptionKey []byte // 32-byte AES key
}

func NewMFAUsecase(userRepo interfaces.UserRepo, mfaRepo repository.MFARepository, auditRepo repository.AuditRepository, encryptionKey []byte) MFAUsecase {
	return &mfaUsecase{
		userRepo:      userRepo,
		mfaRepo:       mfaRepo,
		auditRepo:     auditRepo,
		encryptionKey: encryptionKey,
	}
}

func (u *mfaUsecase) encryptSecret(secret string) (string, error) {
	if len(u.encryptionKey) == 0 {
		return secret, nil // fallback: no encryption configured
	}
	return encrypt.Encrypt(secret, u.encryptionKey)
}

func (u *mfaUsecase) decryptSecret(encrypted string) (string, error) {
	if len(u.encryptionKey) == 0 {
		return encrypted, nil // fallback: no encryption configured
	}
	return encrypt.Decrypt(encrypted, u.encryptionKey)
}

func (u *mfaUsecase) Setup(ctx context.Context, userID int64, email string) (*MFASetupResult, error) {
	existing, err := u.mfaRepo.GetByUserID(ctx, userID)
	if err == nil && existing.IsEnabled() {
		return nil, ErrMFAAlreadyEnabled
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      TOTPIssuer,
		AccountName: email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	encryptedSecret, err := u.encryptSecret(key.Secret())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt MFA secret: %w", err)
	}

	_, err = u.mfaRepo.Create(ctx, userID, encryptedSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to store MFA secret: %w", err)
	}

	return &MFASetupResult{
		Secret:    key.Secret(),
		QRCodeURI: key.URL(),
	}, nil
}

func (u *mfaUsecase) Enable(ctx context.Context, userID int64, code string) ([]string, error) {
	mfa, err := u.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, ErrMFASetupRequired
	}

	if mfa.IsEnabled() {
		return nil, ErrMFAAlreadyEnabled
	}

	secret, err := u.decryptSecret(mfa.SecretEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt MFA secret: %w", err)
	}

	if !totp.Validate(code, secret) {
		return nil, ErrInvalidTOTPCode
	}

	backupCodes := generateBackupCodes(BackupCodeCount, BackupCodeLength)
	hashedCodes := hashBackupCodes(backupCodes)

	if err := u.mfaRepo.Enable(ctx, userID, hashedCodes); err != nil {
		return nil, fmt.Errorf("failed to enable MFA: %w", err)
	}

	u.logAudit(ctx, userID, entity.AuditActionUserMFAEnabled)

	return backupCodes, nil
}

func (u *mfaUsecase) Disable(ctx context.Context, userID int64, code string) error {
	mfa, err := u.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return ErrMFANotEnabled
	}

	if !mfa.IsEnabled() {
		return ErrMFANotEnabled
	}

	secret, err := u.decryptSecret(mfa.SecretEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt MFA secret: %w", err)
	}

	if !totp.Validate(code, secret) {
		if !u.verifyBackupCodeInternal(mfa.BackupCodesHash, code) {
			return ErrInvalidTOTPCode
		}
	}

	if err := u.mfaRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to disable MFA: %w", err)
	}

	u.logAudit(ctx, userID, entity.AuditActionUserMFADisabled)

	return nil
}

func (u *mfaUsecase) Verify(ctx context.Context, userID int64, code string) error {
	mfa, err := u.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return ErrMFANotEnabled
	}

	if !mfa.IsEnabled() {
		return ErrMFANotEnabled
	}

	secret, err := u.decryptSecret(mfa.SecretEncrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt MFA secret: %w", err)
	}

	if !totp.Validate(code, secret) {
		return ErrInvalidTOTPCode
	}

	_ = u.mfaRepo.UpdateLastUsed(ctx, userID)

	return nil
}

func (u *mfaUsecase) VerifyBackupCode(ctx context.Context, userID int64, code string) error {
	mfa, err := u.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return ErrMFANotEnabled
	}

	if !mfa.IsEnabled() {
		return ErrMFANotEnabled
	}

	codeHash := hashSingleBackupCode(code)
	foundIdx := -1
	for i, h := range mfa.BackupCodesHash {
		if subtle.ConstantTimeCompare([]byte(h), []byte(codeHash)) == 1 {
			foundIdx = i
			break
		}
	}

	if foundIdx == -1 {
		return ErrInvalidBackupCode
	}

	newCodes := append(mfa.BackupCodesHash[:foundIdx], mfa.BackupCodesHash[foundIdx+1:]...)
	if err := u.mfaRepo.UpdateBackupCodes(ctx, userID, newCodes); err != nil {
		return fmt.Errorf("failed to update backup codes: %w", err)
	}

	_ = u.mfaRepo.UpdateLastUsed(ctx, userID)

	return nil
}

func (u *mfaUsecase) RegenerateBackupCodes(ctx context.Context, userID int64, code string) ([]string, error) {
	mfa, err := u.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, ErrMFANotEnabled
	}

	if !mfa.IsEnabled() {
		return nil, ErrMFANotEnabled
	}

	secret, err := u.decryptSecret(mfa.SecretEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt MFA secret: %w", err)
	}

	if !totp.Validate(code, secret) {
		return nil, ErrInvalidTOTPCode
	}

	backupCodes := generateBackupCodes(BackupCodeCount, BackupCodeLength)
	hashedCodes := hashBackupCodes(backupCodes)

	if err := u.mfaRepo.UpdateBackupCodes(ctx, userID, hashedCodes); err != nil {
		return nil, fmt.Errorf("failed to update backup codes: %w", err)
	}

	return backupCodes, nil
}

func (u *mfaUsecase) IsEnabled(ctx context.Context, userID int64) (bool, error) {
	mfa, err := u.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return mfa.IsEnabled(), nil
}

func (u *mfaUsecase) verifyBackupCodeInternal(hashedCodes []string, code string) bool {
	codeHash := hashSingleBackupCode(code)
	for _, h := range hashedCodes {
		if subtle.ConstantTimeCompare([]byte(h), []byte(codeHash)) == 1 {
			return true
		}
	}
	return false
}

func (u *mfaUsecase) logAudit(ctx context.Context, userID int64, action string) {
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

func generateBackupCodes(count, length int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		codes[i] = generateRandomCode(length)
	}
	return codes
}

func generateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

func hashBackupCodes(codes []string) []string {
	hashed := make([]string, len(codes))
	for i, code := range codes {
		hashed[i] = hashSingleBackupCode(code)
	}
	return hashed
}

func hashSingleBackupCode(code string) string {
	h := sha256.Sum256([]byte(code))
	return hex.EncodeToString(h[:])
}

type MFALoginResult struct {
	MFARequired bool   `json:"mfa_required"`
	MFAToken    string `json:"mfa_token,omitempty"`
}

func (u *mfaUsecase) GenerateMFAToken(userID int64, secret string, expiry time.Duration) (string, error) {
	return "", nil
}

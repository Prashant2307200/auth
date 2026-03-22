package repository

import (
	"context"
	"database/sql"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/lib/pq"
)

type MFARepository interface {
	Create(ctx context.Context, userID int64, secretEncrypted string) (*entity.UserMFA, error)
	GetByUserID(ctx context.Context, userID int64) (*entity.UserMFA, error)
	Enable(ctx context.Context, userID int64, backupCodesHash []string) error
	Disable(ctx context.Context, userID int64) error
	UpdateBackupCodes(ctx context.Context, userID int64, backupCodesHash []string) error
	UpdateLastUsed(ctx context.Context, userID int64) error
	Delete(ctx context.Context, userID int64) error
}

type mfaRepo struct {
	db *sql.DB
}

func NewMFARepo(db *sql.DB) MFARepository {
	return &mfaRepo{db: db}
}

func (r *mfaRepo) Create(ctx context.Context, userID int64, secretEncrypted string) (*entity.UserMFA, error) {
	query := `
		INSERT INTO user_mfa (user_id, secret_encrypted)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET secret_encrypted = $2, enabled_at = NULL, created_at = NOW()
		RETURNING id, user_id, secret_encrypted, backup_codes_hash, enabled_at, last_used_at, created_at
	`
	var mfa entity.UserMFA
	err := r.db.QueryRowContext(ctx, query, userID, secretEncrypted).Scan(
		&mfa.ID,
		&mfa.UserID,
		&mfa.SecretEncrypted,
		pq.Array(&mfa.BackupCodesHash),
		&mfa.EnabledAt,
		&mfa.LastUsedAt,
		&mfa.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &mfa, nil
}

func (r *mfaRepo) GetByUserID(ctx context.Context, userID int64) (*entity.UserMFA, error) {
	query := `
		SELECT id, user_id, secret_encrypted, backup_codes_hash, enabled_at, last_used_at, created_at
		FROM user_mfa
		WHERE user_id = $1
	`
	var mfa entity.UserMFA
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&mfa.ID,
		&mfa.UserID,
		&mfa.SecretEncrypted,
		pq.Array(&mfa.BackupCodesHash),
		&mfa.EnabledAt,
		&mfa.LastUsedAt,
		&mfa.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &mfa, nil
}

func (r *mfaRepo) Enable(ctx context.Context, userID int64, backupCodesHash []string) error {
	query := `UPDATE user_mfa SET enabled_at = NOW(), backup_codes_hash = $2 WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, pq.Array(backupCodesHash))
	return err
}

func (r *mfaRepo) Disable(ctx context.Context, userID int64) error {
	query := `UPDATE user_mfa SET enabled_at = NULL, backup_codes_hash = NULL WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *mfaRepo) UpdateBackupCodes(ctx context.Context, userID int64, backupCodesHash []string) error {
	query := `UPDATE user_mfa SET backup_codes_hash = $2 WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, pq.Array(backupCodesHash))
	return err
}

func (r *mfaRepo) UpdateLastUsed(ctx context.Context, userID int64) error {
	query := `UPDATE user_mfa SET last_used_at = NOW() WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *mfaRepo) Delete(ctx context.Context, userID int64) error {
	query := `DELETE FROM user_mfa WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

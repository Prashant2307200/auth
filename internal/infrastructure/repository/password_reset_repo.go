package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
)

type PasswordResetRepository interface {
	Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (*entity.PasswordResetToken, error)
	FindByHash(ctx context.Context, tokenHash string) (*entity.PasswordResetToken, error)
	MarkUsed(ctx context.Context, id int64) error
	DeleteAllForUser(ctx context.Context, userID int64) error
}

type passwordResetRepo struct {
	db *sql.DB
}

func NewPasswordResetRepo(db *sql.DB) PasswordResetRepository {
	return &passwordResetRepo{db: db}
}

func (r *passwordResetRepo) Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (*entity.PasswordResetToken, error) {
	query := `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, used_at, created_at
	`
	var token entity.PasswordResetToken
	err := r.db.QueryRowContext(ctx, query, userID, tokenHash, expiresAt).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *passwordResetRepo) FindByHash(ctx context.Context, tokenHash string) (*entity.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`
	var token entity.PasswordResetToken
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.UsedAt,
		&token.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *passwordResetRepo) MarkUsed(ctx context.Context, id int64) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *passwordResetRepo) DeleteAllForUser(ctx context.Context, userID int64) error {
	query := `DELETE FROM password_reset_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

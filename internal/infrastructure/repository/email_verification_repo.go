package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
)

type EmailVerificationRepository interface {
	Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (*entity.EmailVerificationToken, error)
	FindByHash(ctx context.Context, tokenHash string) (*entity.EmailVerificationToken, error)
	DeleteAllForUser(ctx context.Context, userID int64) error
}

type emailVerificationRepo struct {
	db *sql.DB
}

func NewEmailVerificationRepo(db *sql.DB) EmailVerificationRepository {
	return &emailVerificationRepo{db: db}
}

func (r *emailVerificationRepo) Create(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (*entity.EmailVerificationToken, error) {
	query := `
		INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, created_at
	`
	var token entity.EmailVerificationToken
	err := r.db.QueryRowContext(ctx, query, userID, tokenHash, expiresAt).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *emailVerificationRepo) FindByHash(ctx context.Context, tokenHash string) (*entity.EmailVerificationToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM email_verification_tokens
		WHERE token_hash = $1
	`
	var token entity.EmailVerificationToken
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *emailVerificationRepo) DeleteAllForUser(ctx context.Context, userID int64) error {
	query := `DELETE FROM email_verification_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

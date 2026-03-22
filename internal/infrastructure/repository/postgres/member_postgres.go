package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/pkg/db"
)

type MemberPostgres struct {
	Db *sql.DB
}

func NewMemberPostgres(database *sql.DB) (*MemberPostgres, error) {
	if database == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return &MemberPostgres{Db: database}, nil
}

func (m *MemberPostgres) Create(ctx context.Context, member *entity.BusinessMember) error {
	if member == nil {
		return fmt.Errorf("member cannot be nil")
	}
	q := `INSERT INTO business_members (business_id, user_id, email, role_id, status, invited_by, invited_at, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) RETURNING id, created_at, updated_at`
	row, err := db.QueryRow(ctx, m.Db, q, member.BusinessID, member.UserID, member.Email, member.RoleID, member.Status, member.InvitedBy, member.InvitedAt)
	if err != nil {
		return fmt.Errorf("failed to create member: %w", err)
	}
	var id int64
	var created, updated sql.NullTime
	if err := row.Scan(&id, &created, &updated); err != nil {
		return fmt.Errorf("failed to scan created member: %w", err)
	}
	member.ID = id
	if created.Valid {
		member.CreatedAt = created.Time
	}
	if updated.Valid {
		member.UpdatedAt = updated.Time
	}
	return nil
}

func (m *MemberPostgres) GetByID(ctx context.Context, id int64) (*entity.BusinessMember, error) {
	q := `SELECT id, business_id, user_id, email, role_id, status, invited_by, invited_at, accepted_at, created_at, updated_at FROM business_members WHERE id = $1`
	row, err := db.QueryRow(ctx, m.Db, q, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query member: %w", err)
	}
	member := &entity.BusinessMember{}
	var userID sql.NullInt64
	var acceptedAt sql.NullTime
	if err := row.Scan(&member.ID, &member.BusinessID, &userID, &member.Email, &member.RoleID, &member.Status, &member.InvitedBy, &member.InvitedAt, &acceptedAt, &member.CreatedAt, &member.UpdatedAt); err != nil {
		return nil, db.HandleNotFoundError(err, "member", id)
	}
	if userID.Valid {
		uid := userID.Int64
		member.UserID = &uid
	}
	if acceptedAt.Valid {
		member.AcceptedAt = &acceptedAt.Time
	}
	return member, nil
}

func (m *MemberPostgres) GetByUserAndBusiness(ctx context.Context, userID, businessID int64) (*entity.BusinessMember, error) {
	q := `SELECT id, business_id, user_id, email, role_id, status, invited_by, invited_at, accepted_at, created_at, updated_at FROM business_members WHERE user_id = $1 AND business_id = $2`
	row, err := db.QueryRow(ctx, m.Db, q, userID, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to query member: %w", err)
	}
	member := &entity.BusinessMember{}
	var uID sql.NullInt64
	var acceptedAt sql.NullTime
	if err := row.Scan(&member.ID, &member.BusinessID, &uID, &member.Email, &member.RoleID, &member.Status, &member.InvitedBy, &member.InvitedAt, &acceptedAt, &member.CreatedAt, &member.UpdatedAt); err != nil {
		return nil, db.HandleNotFoundError(err, "member", fmt.Sprintf("user:%d business:%d", userID, businessID))
	}
	if uID.Valid {
		uid := uID.Int64
		member.UserID = &uid
	}
	if acceptedAt.Valid {
		member.AcceptedAt = &acceptedAt.Time
	}
	return member, nil
}

func (m *MemberPostgres) ListByBusiness(ctx context.Context, businessID int64) ([]*entity.BusinessMember, error) {
	q := `SELECT id, business_id, user_id, email, role_id, status, invited_by, invited_at, accepted_at, created_at, updated_at FROM business_members WHERE business_id = $1 ORDER BY created_at ASC`
	rows, err := db.QueryRows(ctx, m.Db, q, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	defer rows.Close()
	var out []*entity.BusinessMember
	for rows.Next() {
		var mbr entity.BusinessMember
		var userID sql.NullInt64
		var acceptedAt sql.NullTime
		if err := rows.Scan(&mbr.ID, &mbr.BusinessID, &userID, &mbr.Email, &mbr.RoleID, &mbr.Status, &mbr.InvitedBy, &mbr.InvitedAt, &acceptedAt, &mbr.CreatedAt, &mbr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		if userID.Valid {
			uid := userID.Int64
			mbr.UserID = &uid
		}
		if acceptedAt.Valid {
			mbr.AcceptedAt = &acceptedAt.Time
		}
		out = append(out, &mbr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return out, nil
}

func (m *MemberPostgres) ListByUser(ctx context.Context, userID int64) ([]*entity.BusinessMember, error) {
	q := `SELECT id, business_id, user_id, email, role_id, status, invited_by, invited_at, accepted_at, created_at, updated_at FROM business_members WHERE user_id = $1 ORDER BY created_at ASC`
	rows, err := db.QueryRows(ctx, m.Db, q, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	defer rows.Close()
	var out []*entity.BusinessMember
	for rows.Next() {
		var mbr entity.BusinessMember
		var userIDNull sql.NullInt64
		var acceptedAt sql.NullTime
		if err := rows.Scan(&mbr.ID, &mbr.BusinessID, &userIDNull, &mbr.Email, &mbr.RoleID, &mbr.Status, &mbr.InvitedBy, &mbr.InvitedAt, &acceptedAt, &mbr.CreatedAt, &mbr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		if userIDNull.Valid {
			uid := userIDNull.Int64
			mbr.UserID = &uid
		}
		if acceptedAt.Valid {
			mbr.AcceptedAt = &acceptedAt.Time
		}
		out = append(out, &mbr)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return out, nil
}

func (m *MemberPostgres) GetByInviteToken(ctx context.Context, token string) (*entity.BusinessMember, error) {
	q := `SELECT id, business_id, user_id, email, role_id, status, invited_by, invited_at, accepted_at, invite_token, token_expires_at, created_at, updated_at FROM business_members WHERE invite_token = $1`
	row, err := db.QueryRow(ctx, m.Db, q, token)
	if err != nil {
		return nil, fmt.Errorf("failed to query member by invite token: %w", err)
	}
	member := &entity.BusinessMember{}
	var userIDNull sql.NullInt64
	var acceptedAt sql.NullTime
	var tokenExpires sql.NullTime
	var inviteToken sql.NullString
	if err := row.Scan(&member.ID, &member.BusinessID, &userIDNull, &member.Email, &member.RoleID, &member.Status, &member.InvitedBy, &member.InvitedAt, &acceptedAt, &inviteToken, &tokenExpires, &member.CreatedAt, &member.UpdatedAt); err != nil {
		return nil, db.HandleNotFoundError(err, "member", token)
	}
	if userIDNull.Valid {
		uid := userIDNull.Int64
		member.UserID = &uid
	}
	if acceptedAt.Valid {
		member.AcceptedAt = &acceptedAt.Time
	}
	if inviteToken.Valid {
		member.InviteToken = inviteToken.String
	}
	if tokenExpires.Valid {
		member.TokenExpiresAt = &tokenExpires.Time
	}
	return member, nil
}

func (m *MemberPostgres) Update(ctx context.Context, member *entity.BusinessMember) error {
	if member == nil {
		return fmt.Errorf("member cannot be nil")
	}
	q := `UPDATE business_members SET user_id = $1, email = $2, role_id = $3, status = $4, invited_by = $5, invited_at = $6, accepted_at = $7, updated_at = NOW() WHERE id = $8 AND business_id = $9`
	res, err := db.Exec(ctx, m.Db, q, member.UserID, member.Email, member.RoleID, member.Status, member.InvitedBy, member.InvitedAt, member.AcceptedAt, member.ID, member.BusinessID)
	if err != nil {
		return fmt.Errorf("failed to update member: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return db.HandleNotFoundError(sql.ErrNoRows, "member", member.ID)
	}
	return nil
}

func (m *MemberPostgres) Delete(ctx context.Context, id int64) error {
	q := `DELETE FROM business_members WHERE id = $1`
	res, err := db.Exec(ctx, m.Db, q, id)
	if err != nil {
		return fmt.Errorf("failed to delete member: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return db.HandleNotFoundError(sql.ErrNoRows, "member", id)
	}
	return nil
}

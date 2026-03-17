package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
	"github.com/Prashant2307200/auth-service/pkg/db"
)

type BusinessRepo struct {
	Db *sql.DB
}

func NewBusinessRepo(database *sql.DB) (interfaces.BusinessRepo, error) {
	if database == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}
	return &BusinessRepo{Db: database}, nil
}

func (r *BusinessRepo) Create(ctx context.Context, business *entity.Business) (int64, error) {
	signupPolicy := business.SignupPolicy
	if signupPolicy == "" {
		signupPolicy = entity.SignupPolicyClosed
	}
	query := `
		INSERT INTO businesses (name, slug, email, owner_id, signup_policy, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`
	row, err := db.QueryRow(ctx, r.Db, query, business.Name, business.Slug, business.Email, business.OwnerID, signupPolicy)
	if err != nil {
		return 0, fmt.Errorf("failed to create business: %w", err)
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("failed to get created business ID: %w", err)
	}
	return id, nil
}

func (r *BusinessRepo) GetById(ctx context.Context, id int64) (*entity.Business, error) {
	query := `
		SELECT id, name, slug, email, owner_id, COALESCE(signup_policy, 'closed'), created_at, updated_at
		FROM businesses
		WHERE id = $1
	`
	row, err := db.QueryRow(ctx, r.Db, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query business by id: %w", err)
	}

	var business entity.Business
	if err := row.Scan(&business.ID, &business.Name, &business.Slug, &business.Email, &business.OwnerID, &business.SignupPolicy, &business.CreatedAt, &business.UpdatedAt); err != nil {
		return nil, db.HandleNotFoundError(err, "business", id)
	}
	return &business, nil
}

func (r *BusinessRepo) GetBySlug(ctx context.Context, slug string) (*entity.Business, error) {
	query := `
		SELECT id, name, slug, email, owner_id, COALESCE(signup_policy, 'closed'), created_at, updated_at
		FROM businesses
		WHERE slug = $1
	`
	row, err := db.QueryRow(ctx, r.Db, query, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to query business by slug: %w", err)
	}

	var business entity.Business
	if err := row.Scan(&business.ID, &business.Name, &business.Slug, &business.Email, &business.OwnerID, &business.SignupPolicy, &business.CreatedAt, &business.UpdatedAt); err != nil {
		return nil, db.HandleNotFoundError(err, "business", slug)
	}
	return &business, nil
}

func (r *BusinessRepo) GetByOwnerId(ctx context.Context, ownerId int64) ([]*entity.Business, error) {
	query := `
		SELECT id, name, slug, email, owner_id, COALESCE(signup_policy, 'closed'), created_at, updated_at
		FROM businesses
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`
	rows, err := db.QueryRows(ctx, r.Db, query, ownerId)
	if err != nil {
		return nil, fmt.Errorf("failed to get businesses by owner: %w", err)
	}
	defer rows.Close()

	var businesses []*entity.Business
	for rows.Next() {
		var business entity.Business
		if err := rows.Scan(&business.ID, &business.Name, &business.Slug, &business.Email, &business.OwnerID, &business.SignupPolicy, &business.CreatedAt, &business.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan business: %w", err)
		}
		businesses = append(businesses, &business)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating business rows: %w", err)
	}
	return businesses, nil
}

func (r *BusinessRepo) Update(ctx context.Context, id int64, business *entity.Business) error {
	signupPolicy := business.SignupPolicy
	if signupPolicy == "" {
		signupPolicy = entity.SignupPolicyClosed
	}
	query := `
		UPDATE businesses
		SET name = $1, slug = $2, email = $3, signup_policy = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`
	res, err := db.Exec(ctx, r.Db, query, business.Name, business.Slug, business.Email, signupPolicy, id)
	if err != nil {
		return fmt.Errorf("failed to update business: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return db.HandleNotFoundError(sql.ErrNoRows, "business", id)
	}
	return nil
}

func (r *BusinessRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM businesses WHERE id = $1`
	res, err := db.Exec(ctx, r.Db, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete business: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return db.HandleNotFoundError(sql.ErrNoRows, "business", id)
	}
	return nil
}

func (r *BusinessRepo) List(ctx context.Context) ([]*entity.Business, error) {
	query := `
		SELECT id, name, slug, email, owner_id, COALESCE(signup_policy, 'closed'), created_at, updated_at
		FROM businesses
		ORDER BY created_at DESC
	`
	rows, err := db.QueryRows(ctx, r.Db, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list businesses: %w", err)
	}
	defer rows.Close()

	var businesses []*entity.Business
	for rows.Next() {
		var business entity.Business
		if err := rows.Scan(&business.ID, &business.Name, &business.Slug, &business.Email, &business.OwnerID, &business.SignupPolicy, &business.CreatedAt, &business.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan business: %w", err)
		}
		businesses = append(businesses, &business)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating business rows: %w", err)
	}
	return businesses, nil
}

func (r *BusinessRepo) AddUser(ctx context.Context, businessID int64, userID int64, role int) error {
	query := `
		INSERT INTO business_users (business_id, user_id, role, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
	`
	if _, err := db.Exec(ctx, r.Db, query, businessID, userID, role); err != nil {
		return fmt.Errorf("failed to add user to business: %w", err)
	}
	return nil
}

func (r *BusinessRepo) RemoveUser(ctx context.Context, businessID int64, userID int64) error {
	query := `
		DELETE FROM business_users
		WHERE business_id = $1 AND user_id = $2
	`
	res, err := db.Exec(ctx, r.Db, query, businessID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from business: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return db.HandleNotFoundError(sql.ErrNoRows, "business user", fmt.Sprintf("%d:%d", businessID, userID))
	}
	return nil
}

func (r *BusinessRepo) GetUsers(ctx context.Context, businessID int64) ([]*entity.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password, u.profile_pic, u.role, u.created_at, u.updated_at
		FROM business_users bu
		INNER JOIN users u ON u.id = bu.user_id
		WHERE bu.business_id = $1
		ORDER BY u.username
	`
	rows, err := db.QueryRows(ctx, r.Db, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.ProfilePic, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan business user: %w", err)
		}
		businessIDCopy := businessID
		user.BusinessID = &businessIDCopy
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating business user rows: %w", err)
	}
	return users, nil
}

func (r *BusinessRepo) AddUserIfNotExists(ctx context.Context, businessID int64, userID int64, role int) error {
	query := `
		INSERT INTO business_users (business_id, user_id, role, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (business_id, user_id) DO NOTHING
	`
	_, err := db.Exec(ctx, r.Db, query, businessID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to add user to business: %w", err)
	}
	return nil
}

func (r *BusinessRepo) HasMembership(ctx context.Context, businessID int64, userID int64) (bool, error) {
	query := `
		SELECT 1 FROM business_users WHERE business_id = $1 AND user_id = $2
	`
	row, err := db.QueryRow(ctx, r.Db, query, businessID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check membership: %w", err)
	}
	var dummy int
	if err := row.Scan(&dummy); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to scan membership: %w", err)
	}
	return true, nil
}

func (r *BusinessRepo) GetUserBusinesses(ctx context.Context, userID int64) ([]*entity.Business, error) {
	query := `
		SELECT b.id, b.name, b.slug, b.email, b.owner_id, COALESCE(b.signup_policy, 'closed'), b.created_at, b.updated_at
		FROM business_users bu
		INNER JOIN businesses b ON b.id = bu.business_id
		WHERE bu.user_id = $1
		ORDER BY b.created_at DESC
	`
	rows, err := db.QueryRows(ctx, r.Db, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user businesses: %w", err)
	}
	defer rows.Close()

	var businesses []*entity.Business
	for rows.Next() {
		var business entity.Business
		if err := rows.Scan(&business.ID, &business.Name, &business.Slug, &business.Email, &business.OwnerID, &business.SignupPolicy, &business.CreatedAt, &business.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user business: %w", err)
		}
		businesses = append(businesses, &business)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user business rows: %w", err)
	}
	return businesses, nil
}

func (r *BusinessRepo) GetUserRole(ctx context.Context, businessID int64, userID int64) (int, error) {
	query := `
		SELECT role
		FROM business_users
		WHERE business_id = $1 AND user_id = $2
	`
	row, err := db.QueryRow(ctx, r.Db, query, businessID, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to query user role: %w", err)
	}
	var role int
	if err := row.Scan(&role); err != nil {
		if err == sql.ErrNoRows {
			return 0, err
		}
		return 0, fmt.Errorf("failed to scan user role: %w", err)
	}
	return role, nil
}

func (r *BusinessRepo) CreateInvite(ctx context.Context, invite *entity.BusinessInvite) (int64, error) {
	query := `
		INSERT INTO business_invites (business_id, email, role, invited_by, token, expires_at, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
		RETURNING id
	`
	status := invite.Status
	if status == "" {
		status = entity.InviteStatusPending
	}
	row, err := db.QueryRow(ctx, r.Db, query, invite.BusinessID, invite.Email, invite.Role, invite.InvitedBy, invite.Token, invite.ExpiresAt, status)
	if err != nil {
		return 0, fmt.Errorf("failed to create invite: %w", err)
	}
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("failed to get created invite ID: %w", err)
	}
	return id, nil
}

func (r *BusinessRepo) GetInviteByToken(ctx context.Context, token string) (*entity.BusinessInvite, error) {
	query := `
		SELECT id, business_id, email, role, invited_by, token, expires_at, status, accepted_at, created_at
		FROM business_invites
		WHERE token = $1
	`
	row, err := db.QueryRow(ctx, r.Db, query, token)
	if err != nil {
		return nil, fmt.Errorf("failed to query invite by token: %w", err)
	}
	var inv entity.BusinessInvite
	if err := row.Scan(&inv.ID, &inv.BusinessID, &inv.Email, &inv.Role, &inv.InvitedBy, &inv.Token, &inv.ExpiresAt, &inv.Status, &inv.AcceptedAt, &inv.CreatedAt); err != nil {
		return nil, db.HandleNotFoundError(err, "invite", token)
	}
	return &inv, nil
}

func (r *BusinessRepo) RevokeInvite(ctx context.Context, inviteID int64, businessID int64) error {
	query := `UPDATE business_invites SET status = $1 WHERE id = $2 AND business_id = $3 AND status = $4`
	res, err := db.Exec(ctx, r.Db, query, entity.InviteStatusRevoked, inviteID, businessID, entity.InviteStatusPending)
	if err != nil {
		return fmt.Errorf("failed to revoke invite: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("invite not found or already used")
	}
	return nil
}

func (r *BusinessRepo) AcceptInvite(ctx context.Context, inviteID int64) error {
	query := `UPDATE business_invites SET status = $1, accepted_at = CURRENT_TIMESTAMP WHERE id = $2 AND status = $3`
	res, err := db.Exec(ctx, r.Db, query, entity.InviteStatusAccepted, inviteID, entity.InviteStatusPending)
	if err != nil {
		return fmt.Errorf("failed to accept invite: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("invite not found or already used")
	}
	return nil
}

func (r *BusinessRepo) ListInvites(ctx context.Context, businessID int64) ([]*entity.BusinessInvite, error) {
	query := `
		SELECT id, business_id, email, role, invited_by, token, expires_at, status, accepted_at, created_at
		FROM business_invites
		WHERE business_id = $1
		ORDER BY created_at DESC
	`
	rows, err := db.QueryRows(ctx, r.Db, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invites: %w", err)
	}
	defer rows.Close()
	var list []*entity.BusinessInvite
	for rows.Next() {
		var inv entity.BusinessInvite
		if err := rows.Scan(&inv.ID, &inv.BusinessID, &inv.Email, &inv.Role, &inv.InvitedBy, &inv.Token, &inv.ExpiresAt, &inv.Status, &inv.AcceptedAt, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan invite: %w", err)
		}
		list = append(list, &inv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating invite rows: %w", err)
	}
	return list, nil
}

func (r *BusinessRepo) CreateDomain(ctx context.Context, domain *entity.BusinessDomain) (int64, error) {
	query := `
		INSERT INTO business_domains (business_id, domain, verified, auto_join_enabled, verification_token, verified_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		RETURNING id
	`
	var verifiedAt interface{}
	if domain.VerifiedAt != nil {
		verifiedAt = domain.VerifiedAt
	}
	row, err := db.QueryRow(ctx, r.Db, query, domain.BusinessID, domain.Domain, domain.Verified, domain.AutoJoinEnabled, domain.VerificationToken, verifiedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to create domain: %w", err)
	}
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("failed to get created domain ID: %w", err)
	}
	return id, nil
}

func (r *BusinessRepo) GetDomain(ctx context.Context, businessID int64, domain string) (*entity.BusinessDomain, error) {
	query := `
		SELECT id, business_id, domain, verified, auto_join_enabled, verification_token, verified_at, created_at
		FROM business_domains
		WHERE business_id = $1 AND domain = $2
	`
	row, err := db.QueryRow(ctx, r.Db, query, businessID, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to query domain: %w", err)
	}
	var d entity.BusinessDomain
	if err := row.Scan(&d.ID, &d.BusinessID, &d.Domain, &d.Verified, &d.AutoJoinEnabled, &d.VerificationToken, &d.VerifiedAt, &d.CreatedAt); err != nil {
		return nil, db.HandleNotFoundError(err, "domain", domain)
	}
	return &d, nil
}

func (r *BusinessRepo) GetDomainByVerificationToken(ctx context.Context, token string) (*entity.BusinessDomain, error) {
	query := `
		SELECT id, business_id, domain, verified, auto_join_enabled, verification_token, verified_at, created_at
		FROM business_domains
		WHERE verification_token = $1
	`
	row, err := db.QueryRow(ctx, r.Db, query, token)
	if err != nil {
		return nil, fmt.Errorf("failed to query domain by token: %w", err)
	}
	var d entity.BusinessDomain
	if err := row.Scan(&d.ID, &d.BusinessID, &d.Domain, &d.Verified, &d.AutoJoinEnabled, &d.VerificationToken, &d.VerifiedAt, &d.CreatedAt); err != nil {
		return nil, db.HandleNotFoundError(err, "domain", token)
	}
	return &d, nil
}

func (r *BusinessRepo) FindAutoJoinBusinessByEmailDomain(ctx context.Context, emailDomain string) (*entity.Business, error) {
	query := `
		SELECT b.id, b.name, b.slug, b.email, b.owner_id, COALESCE(b.signup_policy, 'closed'), b.created_at, b.updated_at
		FROM business_domains bd
		INNER JOIN businesses b ON b.id = bd.business_id
		WHERE LOWER(bd.domain) = LOWER($1) AND bd.verified = true AND bd.auto_join_enabled = true
		LIMIT 1
	`
	row, err := db.QueryRow(ctx, r.Db, query, emailDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to find auto-join business: %w", err)
	}
	var business entity.Business
	if err := row.Scan(&business.ID, &business.Name, &business.Slug, &business.Email, &business.OwnerID, &business.SignupPolicy, &business.CreatedAt, &business.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, db.HandleNotFoundError(err, "business", emailDomain)
	}
	return &business, nil
}

func (r *BusinessRepo) VerifyDomain(ctx context.Context, domainID int64) error {
	query := `UPDATE business_domains SET verified = true, verified_at = CURRENT_TIMESTAMP, verification_token = NULL WHERE id = $1`
	res, err := db.Exec(ctx, r.Db, query, domainID)
	if err != nil {
		return fmt.Errorf("failed to verify domain: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("domain not found")
	}
	return nil
}

func (r *BusinessRepo) UpdateDomainAutoJoin(ctx context.Context, domainID int64, businessID int64, enabled bool) error {
	query := `UPDATE business_domains SET auto_join_enabled = $1 WHERE id = $2 AND business_id = $3`
	res, err := db.Exec(ctx, r.Db, query, enabled, domainID, businessID)
	if err != nil {
		return fmt.Errorf("failed to update domain auto-join: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("domain not found")
	}
	return nil
}

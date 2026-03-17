package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
	"github.com/Prashant2307200/auth-service/pkg/db"
)

type UserRepo struct {
	Db *sql.DB
}

// NewUserRepo creates a new user repository
// Note: Migrations should be run separately during application startup, not here
func NewUserRepo(database *sql.DB) (interfaces.UserRepo, error) {
	if database == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}
	return &UserRepo{Db: database}, nil
}

func (r *UserRepo) Search(ctx context.Context, currentID int64, search string) ([]*entity.User, error) {
	// Sanitize search input to prevent SQL injection
	sanitizedSearch := db.SanitizeSearchInput(search, 100)
	
	// Use parameterized query with proper escaping
	query := `
		SELECT id, username, email, profile_pic
		FROM users
		WHERE id != $1
		AND ($2::text = '' OR username ILIKE $3 OR email ILIKE $4)
		ORDER BY username
		LIMIT 20
	`
	
	// Build search pattern with proper escaping
	searchPattern := "%" + sanitizedSearch + "%"

	rows, err := db.QueryRows(ctx, r.Db, query, currentID, sanitizedSearch, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.ProfilePic); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}


func (r *UserRepo) List(ctx context.Context) ([]*entity.User, error) {
	// Only select necessary fields - avoid SELECT * for performance
	query := `
		SELECT id, username, email, password, profile_pic, role, created_at, updated_at 
		FROM users 
		ORDER BY created_at DESC
	`
	rows, err := db.QueryRows(ctx, r.Db, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.ProfilePic,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return users, nil
}

func (r *UserRepo) GetById(ctx context.Context, id int64) (*entity.User, error) {
	query := `
		SELECT id, username, email, password, profile_pic, role, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`
	var user entity.User
	row, err := db.QueryRow(ctx, r.Db, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	err = row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.ProfilePic,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, db.HandleNotFoundError(err, "user", id)
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	query := `
		SELECT id, username, email, password, profile_pic, role, created_at, updated_at 
		FROM users 
		WHERE email = $1
	`
	var user entity.User
	row, err := db.QueryRow(ctx, r.Db, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	err = row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.ProfilePic,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) UpdateById(ctx context.Context, id int64, user *entity.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	query := `
		UPDATE users
		SET username = $1, email = $2, password = $3, profile_pic = $4, role = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`
	
	res, err := db.Exec(ctx, r.Db, query, user.Username, user.Email, user.Password, user.ProfilePic, user.Role, id)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return db.HandleNotFoundError(sql.ErrNoRows, "user", id)
	}
	return nil
}

func (r *UserRepo) DeleteById(ctx context.Context, id int64) error {
	query := "DELETE FROM users WHERE id = $1"
	
	res, err := db.Exec(ctx, r.Db, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return db.HandleNotFoundError(sql.ErrNoRows, "user", id)
	}
	return nil
}

func (r *UserRepo) Create(ctx context.Context, user *entity.User) (int64, error) {
	if user == nil {
		return 0, fmt.Errorf("user cannot be nil")
	}

	query := `
		INSERT INTO users (username, email, password, profile_pic, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	row, err := db.QueryRow(ctx, r.Db, query, user.Username, user.Email, user.Password, user.ProfilePic, user.Role)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("failed to get created user ID: %w", err)
	}

	return id, nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
)

type UserRepo struct {
	Db *sql.DB
}

func NewUserRepo(db *sql.DB) (interfaces.UserRepo, error) {
	
	dropTableQuery := "DROP TABLE IF EXISTS users;"
	if _, err := db.Exec(dropTableQuery); err != nil {
		return nil, err
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(15) NOT NULL CHECK (char_length(username) >= 3),
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL CHECK (char_length(password) >= 6),
		profile_pic TEXT DEFAULT '',
		role INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(createTableQuery); err != nil {
		return nil, err
	}

	return &UserRepo{Db: db}, nil
}

func (r *UserRepo) Search(ctx context.Context, currentID int64, search string) ([]*entity.User, error) {
	query := `
		SELECT id, username, email, profile_pic
		FROM users
		WHERE id != $1
		AND ($2::text = '' OR username ILIKE '%' || $2 || '%' OR email ILIKE '%' || $2 || '%')
		ORDER BY username
		LIMIT 20
	`

	rows, err := r.Db.QueryContext(ctx, query, currentID, search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User
	for rows.Next() {
		var user entity.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.ProfilePic)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}


func (r *UserRepo) List(ctx context.Context) ([]*entity.User, error) {
	query := "SELECT id, username, email, password, profile_pic, role, created_at, updated_at FROM users"
	rows, err := r.Db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
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
			return nil, fmt.Errorf("scan error: %w", err)
		}
		users = append(users, &user)
	}
	return users, nil
}

func (r *UserRepo) GetById(ctx context.Context, id int64) (*entity.User, error) {
	query := "SELECT id, username, email, password, profile_pic, role, created_at, updated_at FROM users WHERE id = $1"
	var user entity.User
	err := r.Db.QueryRowContext(ctx, query, id).Scan(
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
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := "SELECT id, username, email, password, profile_pic, role, created_at, updated_at FROM users WHERE email = $1"
	var user entity.User
	err := r.Db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.ProfilePic,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return &user, err
}

func (r *UserRepo) UpdateById(ctx context.Context, id int64, user *entity.User) error {

	query := `
		UPDATE users
		SET username = $1, email = $2, password = $3, profile_pic = $4, role = $5, updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`
	stmt, err := r.Db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, user.Username, user.Email, user.Password, user.ProfilePic, user.Role, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with id %d", id)
	}
	return nil
}

func (r *UserRepo) DeleteById(ctx context.Context, id int64) error {

	query := "DELETE FROM users WHERE id = $1"
	stmt, err := r.Db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no user found with id %d", id)
	}
	return nil
}

func (r *UserRepo) Create(ctx context.Context, user *entity.User) (int64, error) {
	query := `
		INSERT INTO users (username, email, password, profile_pic, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id
	`

	stmt, err := r.Db.PrepareContext(ctx, query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var id int64
	// Use QueryRowContext on the prepared statement, then Scan the id
	err = stmt.QueryRowContext(ctx, user.Username, user.Email, user.Password, user.ProfilePic, user.Role).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

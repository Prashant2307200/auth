package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/pkg/db"
)

type RolePostgres struct {
	Db *sql.DB
}

func NewRolePostgres(database *sql.DB) (*RolePostgres, error) {
	if database == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return &RolePostgres{Db: database}, nil
}

func (r *RolePostgres) Create(ctx context.Context, role *entity.Role) (int64, error) {
	if role == nil {
		return 0, fmt.Errorf("role cannot be nil")
	}
	// Serialize permissions array to JSON string for storage
	permsJSON, _ := json.Marshal(role.Permissions)
	query := `INSERT INTO roles (business_id, name, permissions, created_at, updated_at) VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id`
	row, err := db.QueryRow(ctx, r.Db, query, role.BusinessID, role.Name, string(permsJSON))
	if err != nil {
		return 0, fmt.Errorf("failed to create role: %w", err)
	}
	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, fmt.Errorf("failed to scan created role id: %w", err)
	}
	return id, nil
}

func (r *RolePostgres) GetByID(ctx context.Context, id int64) (*entity.Role, error) {
	query := `SELECT id, business_id, name, permissions, created_at, updated_at FROM roles WHERE id = $1`
	row, err := db.QueryRow(ctx, r.Db, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query role: %w", err)
	}
	var role entity.Role
	var permsStr string
	if err := row.Scan(&role.ID, &role.BusinessID, &role.Name, &permsStr, &role.CreatedAt, &role.UpdatedAt); err != nil {
		return nil, db.HandleNotFoundError(err, "role", id)
	}
	if len(permsStr) > 0 {
		_ = json.Unmarshal([]byte(permsStr), &role.Permissions)
	}
	return &role, nil
}

func (r *RolePostgres) ListByBusiness(ctx context.Context, businessID int64) ([]*entity.Role, error) {
	query := `SELECT id, business_id, name, permissions, created_at, updated_at FROM roles WHERE business_id = $1 ORDER BY name ASC`
	rows, err := db.QueryRows(ctx, r.Db, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var out []*entity.Role
	for rows.Next() {
		var role entity.Role
		var permsStr string
		if err := rows.Scan(&role.ID, &role.BusinessID, &role.Name, &permsStr, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		if len(permsStr) > 0 {
			_ = json.Unmarshal([]byte(permsStr), &role.Permissions)
		}
		out = append(out, &role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return out, nil
}

func (r *RolePostgres) Update(ctx context.Context, role *entity.Role) error {
	if role == nil {
		return fmt.Errorf("role cannot be nil")
	}
	// serialize permissions to JSON string for storage
	permsJSON, _ := json.Marshal(role.Permissions)
	query := `UPDATE roles SET name = $1, permissions = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND business_id = $4`
	res, err := db.Exec(ctx, r.Db, query, role.Name, string(permsJSON), role.ID, role.BusinessID)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return db.HandleNotFoundError(sql.ErrNoRows, "role", role.ID)
	}
	return nil
}

func (r *RolePostgres) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM roles WHERE id = $1`
	res, err := db.Exec(ctx, r.Db, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return db.HandleNotFoundError(sql.ErrNoRows, "role", id)
	}
	return nil
}

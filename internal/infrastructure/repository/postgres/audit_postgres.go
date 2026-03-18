package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/pkg/db"
)

type AuditPostgres struct {
	Db *sql.DB
}

func NewAuditPostgres(database *sql.DB) (*AuditPostgres, error) {
	if database == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return &AuditPostgres{Db: database}, nil
}

func (a *AuditPostgres) Log(ctx context.Context, audit *entity.AuditLog) error {
	if audit == nil {
		return fmt.Errorf("audit cannot be nil")
	}
	oldJSON, _ := json.Marshal(audit.OldValues)
	newJSON, _ := json.Marshal(audit.NewValues)

	q := `INSERT INTO audit_logs (business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`
	_, err := db.Exec(ctx, a.Db, q, audit.BusinessID, audit.UserID, audit.Action, audit.EntityType, audit.EntityID, oldJSON, newJSON, audit.IPAddress, audit.UserAgent)
	if err != nil {
		return fmt.Errorf("failed to insert audit log: %w", err)
	}
	return nil
}

func (a *AuditPostgres) GetByID(ctx context.Context, id int64) (*entity.AuditLog, error) {
	q := `SELECT id, business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE id = $1`
	row, err := db.QueryRow(ctx, a.Db, q, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit log: %w", err)
	}
	var al entity.AuditLog
	var oldBytes, newBytes []byte
	var entityID sql.NullInt64
	if err := row.Scan(&al.ID, &al.BusinessID, &al.UserID, &al.Action, &al.EntityType, &entityID, &oldBytes, &newBytes, &al.IPAddress, &al.UserAgent, &al.CreatedAt); err != nil {
		return nil, db.HandleNotFoundError(err, "audit_log", id)
	}
	if entityID.Valid {
		al.EntityID = &entityID.Int64
	}
	if len(oldBytes) > 0 {
		json.Unmarshal(oldBytes, &al.OldValues)
	}
	if len(newBytes) > 0 {
		json.Unmarshal(newBytes, &al.NewValues)
	}
	return &al, nil
}

func (a *AuditPostgres) ListByBusiness(ctx context.Context, businessID int64, limit, offset int) ([]*entity.AuditLog, error) {
	q := `SELECT id, business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE business_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := db.QueryRows(ctx, a.Db, q, businessID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()
	var out []*entity.AuditLog
	for rows.Next() {
		var al entity.AuditLog
		var oldBytes, newBytes []byte
		var entityID sql.NullInt64
		if err := rows.Scan(&al.ID, &al.BusinessID, &al.UserID, &al.Action, &al.EntityType, &entityID, &oldBytes, &newBytes, &al.IPAddress, &al.UserAgent, &al.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan audit row: %w", err)
		}
		if entityID.Valid {
			al.EntityID = &entityID.Int64
		}
		if len(oldBytes) > 0 {
			json.Unmarshal(oldBytes, &al.OldValues)
		}
		if len(newBytes) > 0 {
			json.Unmarshal(newBytes, &al.NewValues)
		}
		out = append(out, &al)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return out, nil
}

func (a *AuditPostgres) ListByUser(ctx context.Context, businessID, userID int64, limit, offset int) ([]*entity.AuditLog, error) {
	q := `SELECT id, business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE business_id = $1 AND user_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	rows, err := db.QueryRows(ctx, a.Db, q, businessID, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs by user: %w", err)
	}
	defer rows.Close()
	var out []*entity.AuditLog
	for rows.Next() {
		var al entity.AuditLog
		var oldBytes, newBytes []byte
		var entityID sql.NullInt64
		if err := rows.Scan(&al.ID, &al.BusinessID, &al.UserID, &al.Action, &al.EntityType, &entityID, &oldBytes, &newBytes, &al.IPAddress, &al.UserAgent, &al.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan audit row: %w", err)
		}
		if entityID.Valid {
			al.EntityID = &entityID.Int64
		}
		if len(oldBytes) > 0 {
			json.Unmarshal(oldBytes, &al.OldValues)
		}
		if len(newBytes) > 0 {
			json.Unmarshal(newBytes, &al.NewValues)
		}
		out = append(out, &al)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return out, nil
}

// Export returns ALL audit logs for a business (GDPR export). No pagination intentionally.
func (a *AuditPostgres) Export(ctx context.Context, businessID int64) ([]*entity.AuditLog, error) {
	q := `SELECT id, business_id, user_id, action, entity_type, entity_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE business_id = $1 ORDER BY created_at ASC`
	rows, err := db.QueryRows(ctx, a.Db, q, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to export audit logs: %w", err)
	}
	defer rows.Close()
	var out []*entity.AuditLog
	for rows.Next() {
		var al entity.AuditLog
		var oldBytes, newBytes []byte
		var entityID sql.NullInt64
		if err := rows.Scan(&al.ID, &al.BusinessID, &al.UserID, &al.Action, &al.EntityType, &entityID, &oldBytes, &newBytes, &al.IPAddress, &al.UserAgent, &al.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan audit row: %w", err)
		}
		if entityID.Valid {
			al.EntityID = &entityID.Int64
		}
		if len(oldBytes) > 0 {
			json.Unmarshal(oldBytes, &al.OldValues)
		}
		if len(newBytes) > 0 {
			json.Unmarshal(newBytes, &al.NewValues)
		}
		out = append(out, &al)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return out, nil
}

package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Prashant2307200/auth-service/internal/entity"
	postgresrepo "github.com/Prashant2307200/auth-service/internal/infrastructure/repository/postgres"
)

// AuditRepository defines immutable audit log operations
type AuditRepository interface {
	Log(ctx context.Context, audit *entity.AuditLog) error
	GetByID(ctx context.Context, id int64) (*entity.AuditLog, error)
	ListByBusiness(ctx context.Context, businessID int64, limit, offset int) ([]*entity.AuditLog, error)
	ListByUser(ctx context.Context, businessID, userID int64, limit, offset int) ([]*entity.AuditLog, error)
	ListWithFilter(ctx context.Context, businessID int64, userID *int64, action, fromTime, toTime string, limit, offset int) ([]*entity.AuditLog, error)
	Export(ctx context.Context, businessID int64) ([]*entity.AuditLog, error)
}

// NewAuditRepo returns a Postgres-backed audit repository.
func NewAuditRepo(database *sql.DB) (AuditRepository, error) {
	if database == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return postgresrepo.NewAuditPostgres(database)
}

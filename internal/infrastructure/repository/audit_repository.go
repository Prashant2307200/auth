package repository

import (
	"context"

	"github.com/Prashant2307200/auth-service/internal/entity"
)

// AuditRepository defines immutable audit log operations
type AuditRepository interface {
	Log(ctx context.Context, audit *entity.AuditLog) error
	GetByID(ctx context.Context, id int64) (*entity.AuditLog, error)
	ListByBusiness(ctx context.Context, businessID int64, limit, offset int) ([]*entity.AuditLog, error)
	ListByUser(ctx context.Context, businessID, userID int64, limit, offset int) ([]*entity.AuditLog, error)
	Export(ctx context.Context, businessID int64) ([]*entity.AuditLog, error)
}

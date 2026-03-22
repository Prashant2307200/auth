package repository

import (
	"context"

	"github.com/Prashant2307200/auth-service/internal/entity"
)

// RoleRepository defines CRUD operations for roles scoped to a business (tenant)
type RoleRepository interface {
	Create(ctx context.Context, role *entity.Role) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Role, error)
	ListByBusiness(ctx context.Context, businessID int64) ([]*entity.Role, error)
	Update(ctx context.Context, role *entity.Role) error
	Delete(ctx context.Context, id int64) error
}

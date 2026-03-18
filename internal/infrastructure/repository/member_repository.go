package repository

import (
	"context"

	"github.com/Prashant2307200/auth-service/internal/entity"
)

// MemberRepository defines data access for business members/invites
type MemberRepository interface {
	Create(ctx context.Context, member *entity.BusinessMember) error
	GetByID(ctx context.Context, id int64) (*entity.BusinessMember, error)
	GetByUserAndBusiness(ctx context.Context, userID, businessID int64) (*entity.BusinessMember, error)
	ListByBusiness(ctx context.Context, businessID int64) ([]*entity.BusinessMember, error)
	ListByUser(ctx context.Context, userID int64) ([]*entity.BusinessMember, error)
	Update(ctx context.Context, member *entity.BusinessMember) error
	Delete(ctx context.Context, id int64) error
}

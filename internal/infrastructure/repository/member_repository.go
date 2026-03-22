package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Prashant2307200/auth-service/internal/entity"
	postgresrepo "github.com/Prashant2307200/auth-service/internal/infrastructure/repository/postgres"
)

type MemberRepository interface {
	Create(ctx context.Context, member *entity.BusinessMember) error
	GetByID(ctx context.Context, id int64) (*entity.BusinessMember, error)
	GetByUserAndBusiness(ctx context.Context, userID, businessID int64) (*entity.BusinessMember, error)
	GetByInviteToken(ctx context.Context, token string) (*entity.BusinessMember, error)
	ListByBusiness(ctx context.Context, businessID int64) ([]*entity.BusinessMember, error)
	ListByUser(ctx context.Context, userID int64) ([]*entity.BusinessMember, error)
	Update(ctx context.Context, member *entity.BusinessMember) error
	Delete(ctx context.Context, id int64) error
}

func NewMemberRepo(database *sql.DB) (MemberRepository, error) {
	if database == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return postgresrepo.NewMemberPostgres(database)
}

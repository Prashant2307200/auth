package seeder_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/seeder"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSeedAll_Idempotent(t *testing.T) {
	ctx := context.Background()
	userRepo := &testutil.MockUserRepo{}
	businessRepo := &testutil.MockBusinessRepo{}

	// Simulate "second run" - all records already exist
	for _, u := range seeder.SeedUsersData {
		userRepo.On("GetByEmail", mock.Anything, u.Email).Return(&entity.User{ID: 1, Username: u.Username, Email: u.Email}, nil)
	}
	for _, b := range seeder.SeedBusinessesData {
		businessRepo.On("GetBySlug", mock.Anything, b.Slug).Return(&entity.Business{ID: 1, Slug: b.Slug}, nil)
	}
	for _, i := range seeder.SeedInvitesData {
		businessRepo.On("GetInviteByToken", mock.Anything, i.Token).Return(&entity.BusinessInvite{ID: 1, Token: i.Token}, nil)
	}
	businessRepo.On("GetDomain", mock.Anything, int64(1), "acme.com").Return(&entity.BusinessDomain{ID: 1, Domain: "acme.com"}, nil)

	err := seeder.SeedAll(ctx, userRepo, businessRepo)
	require.NoError(t, err)

	userRepo.AssertExpectations(t)
	businessRepo.AssertExpectations(t)
}

func TestSeedUsers_FirstRun(t *testing.T) {
	ctx := context.Background()
	userRepo := &testutil.MockUserRepo{}

	for _, u := range seeder.SeedUsersData {
		userRepo.On("GetByEmail", mock.Anything, u.Email).Return(nil, sql.ErrNoRows)
		userRepo.On("Create", mock.Anything, mock.MatchedBy(func(usr *entity.User) bool {
			return usr.Username == u.Username && usr.Email == u.Email && usr.Role == u.Role
		})).Return(int64(1), nil)
	}

	err := seeder.SeedUsers(ctx, userRepo)
	require.NoError(t, err)
	userRepo.AssertExpectations(t)
}

func TestSeedUsers_SkipsExisting(t *testing.T) {
	ctx := context.Background()
	userRepo := &testutil.MockUserRepo{}

	existing := &entity.User{ID: 1, Username: "admin", Email: "admin@example.com"}
	userRepo.On("GetByEmail", mock.Anything, "admin@example.com").Return(existing, nil)
	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.Anything, mock.Anything).Return(int64(2), nil)
	userRepo.On("GetByEmail", mock.Anything, "demo@example.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.Anything, mock.Anything).Return(int64(3), nil)

	err := seeder.SeedUsers(ctx, userRepo)
	require.NoError(t, err)
	userRepo.AssertExpectations(t)
}

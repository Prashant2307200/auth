package usecase

import (
	"context"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/testutil"
	pkghash "github.com/Prashant2307200/auth-service/pkg/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogin_RehashOldHash(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	tokenService := new(testutil.MockTokenService)
	cloudService := new(testutil.MockCloudService)

	user := testutil.CreateTestUser()
	// create old bcrypt hash cost 10
	oldHash, _ := pkghash.HashPasswordWithCost("password123", 10)
	user.Password = oldHash

	userRepo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil)
	userRepo.On("UpdatePassword", mock.Anything, user.ID, mock.AnythingOfType("string")).Return(nil)
	tokenService.On("GenerateRefreshToken", user.ID).Return("refresh", nil)
	tokenService.On("StoreRefreshToken", mock.Anything, user.ID, "refresh").Return(nil)
	tokenService.On("GenerateAccessToken", user.ID).Return("access", nil)

	uc := NewAuthUseCase(userRepo, nil, tokenService, cloudService)
	access, refresh, err := uc.LoginUser(context.Background(), user.Email, "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)

	userRepo.AssertCalled(t, "UpdatePassword", mock.Anything, user.ID, mock.AnythingOfType("string"))
}

func TestLogin_NoRehashWhenUpToDate(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	tokenService := new(testutil.MockTokenService)
	cloudService := new(testutil.MockCloudService)

	user := testutil.CreateTestUser()
	// create current bcrypt hash cost 12
	goodHash, _ := pkghash.HashPasswordWithCost("password123", pkghash.CurrentCost)
	user.Password = goodHash

	userRepo.On("GetByEmail", mock.Anything, user.Email).Return(user, nil)
	tokenService.On("GenerateRefreshToken", user.ID).Return("refresh", nil)
	tokenService.On("StoreRefreshToken", mock.Anything, user.ID, "refresh").Return(nil)
	tokenService.On("GenerateAccessToken", user.ID).Return("access", nil)

	uc := NewAuthUseCase(userRepo, nil, tokenService, cloudService)
	access, refresh, err := uc.LoginUser(context.Background(), user.Email, "password123")
	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)

	userRepo.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
}

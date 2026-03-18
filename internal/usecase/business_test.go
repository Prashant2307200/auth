package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBusinessUseCase_CreateBusiness(t *testing.T) {
	businessRepo := new(testutil.MockBusinessRepo)
	userRepo := new(testutil.MockUserRepo)
	uc := NewBusinessUseCase(businessRepo, userRepo)

	input := testutil.CreateTestBusiness()
	input.Name, input.Slug = "Acme", "acme"
	input.ID = 0

	expected := testutil.CreateTestBusinessWithID(11)
	expected.Name, expected.Slug = "Acme", "acme"
	expected.OwnerID = 10

	businessRepo.On("GetBySlug", mock.Anything, "acme").Return(nil, errors.New("not found"))
	businessRepo.On("CreateWithOwner", mock.Anything, mock.AnythingOfType("*entity.Business"), int64(10)).Return(int64(11), nil)
	businessRepo.On("GetById", mock.Anything, int64(11)).Return(expected, nil)

	result, err := uc.CreateBusiness(context.Background(), 10, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(11), result.ID)
	assert.Equal(t, int64(10), result.OwnerID)
	businessRepo.AssertExpectations(t)
}

func TestBusinessUseCase_UpdateBusiness(t *testing.T) {
	businessRepo := new(testutil.MockBusinessRepo)
	userRepo := new(testutil.MockUserRepo)
	uc := NewBusinessUseCase(businessRepo, userRepo)

	update := testutil.CreateTestBusinessWithSignupPolicy("updated", "closed")
	update.Name, update.Email = "Updated", "updated@acme.com"
	businessRepo.On("GetUserRole", mock.Anything, int64(2), int64(1)).Return(BusinessRoleAdmin, nil)
	businessRepo.On("Update", mock.Anything, int64(2), update).Return(nil)

	err := uc.UpdateBusiness(context.Background(), 1, 2, update)
	assert.NoError(t, err)
	businessRepo.AssertExpectations(t)
}

func TestBusinessUseCase_DeleteBusiness_OnlyOwner(t *testing.T) {
	businessRepo := new(testutil.MockBusinessRepo)
	userRepo := new(testutil.MockUserRepo)
	uc := NewBusinessUseCase(businessRepo, userRepo)

	businessRepo.On("GetUserRole", mock.Anything, int64(5), int64(2)).Return(BusinessRoleAdmin, nil)
	err := uc.DeleteBusiness(context.Background(), 2, 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only owner")
}

func TestBusinessUseCase_AddUserToBusiness(t *testing.T) {
	businessRepo := new(testutil.MockBusinessRepo)
	userRepo := new(testutil.MockUserRepo)
	uc := NewBusinessUseCase(businessRepo, userRepo)

	userRepo.On("GetById", mock.Anything, int64(9)).Return(testutil.CreateTestUserWithID(9), nil)
	businessRepo.On("GetUserRole", mock.Anything, int64(3), int64(1)).Return(BusinessRoleOwner, nil)
	businessRepo.On("AddUser", mock.Anything, int64(3), int64(9), BusinessRoleMember).Return(nil)

	err := uc.AddUserToBusiness(context.Background(), 1, 3, 9, BusinessRoleMember)
	assert.NoError(t, err)
	businessRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestBusinessUseCase_GetUserBusinesses(t *testing.T) {
	businessRepo := new(testutil.MockBusinessRepo)
	userRepo := new(testutil.MockUserRepo)
	uc := NewBusinessUseCase(businessRepo, userRepo)

	businesses := []*entity.Business{testutil.CreateTestBusiness()}
	businessRepo.On("GetUserBusinesses", mock.Anything, int64(1)).Return(businesses, nil)

	result, err := uc.GetUserBusinesses(context.Background(), 1)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "Acme Inc", result[0].Name)
	businessRepo.AssertExpectations(t)
}

func TestBusinessUseCase_CreateInvite(t *testing.T) {
	businessRepo := new(testutil.MockBusinessRepo)
	userRepo := new(testutil.MockUserRepo)
	uc := NewBusinessUseCase(businessRepo, userRepo)

	businessRepo.On("GetUserRole", mock.Anything, int64(5), int64(1)).Return(BusinessRoleAdmin, nil)
	businessRepo.On("CreateInvite", mock.Anything, mock.AnythingOfType("*entity.BusinessInvite")).Return(int64(10), nil)

	inv, err := uc.CreateInvite(context.Background(), 1, 5, "newuser@acme.com", BusinessRoleMember)
	require.NoError(t, err)
	require.NotNil(t, inv)
	assert.Equal(t, int64(5), inv.BusinessID)
	assert.Equal(t, "newuser@acme.com", inv.Email)
	assert.Equal(t, entity.InviteStatusPending, inv.Status)
	businessRepo.AssertExpectations(t)
}

func TestBusinessUseCase_ListInvites(t *testing.T) {
	businessRepo := new(testutil.MockBusinessRepo)
	userRepo := new(testutil.MockUserRepo)
	uc := NewBusinessUseCase(businessRepo, userRepo)

	invites := []*entity.BusinessInvite{testutil.CreateTestInvite(5, "a@x.com", "tok", entity.InviteStatusPending, time.Now())}
	businessRepo.On("GetUserRole", mock.Anything, int64(5), int64(1)).Return(BusinessRoleAdmin, nil)
	businessRepo.On("ListInvites", mock.Anything, int64(5)).Return(invites, nil)

	result, err := uc.ListInvites(context.Background(), 1, 5)
	require.NoError(t, err)
	assert.Len(t, result, 1)
	businessRepo.AssertExpectations(t)
}

func TestBusinessUseCase_RevokeInvite(t *testing.T) {
	businessRepo := new(testutil.MockBusinessRepo)
	userRepo := new(testutil.MockUserRepo)
	uc := NewBusinessUseCase(businessRepo, userRepo)

	businessRepo.On("GetUserRole", mock.Anything, int64(5), int64(1)).Return(BusinessRoleAdmin, nil)
	businessRepo.On("RevokeInvite", mock.Anything, int64(3), int64(5)).Return(nil)

	err := uc.RevokeInvite(context.Background(), 1, 5, 3)
	assert.NoError(t, err)
	businessRepo.AssertExpectations(t)
}

func TestBusinessUseCase_AddDomain(t *testing.T) {
	businessRepo := new(testutil.MockBusinessRepo)
	userRepo := new(testutil.MockUserRepo)
	uc := NewBusinessUseCase(businessRepo, userRepo)

	businessRepo.On("GetUserRole", mock.Anything, int64(5), int64(1)).Return(BusinessRoleAdmin, nil)
	businessRepo.On("GetDomain", mock.Anything, int64(5), "acme.com").Return(nil, errors.New("not found"))
	businessRepo.On("CreateDomain", mock.Anything, mock.AnythingOfType("*entity.BusinessDomain")).Return(int64(7), nil)

	d, err := uc.AddDomain(context.Background(), 1, 5, "acme.com")
	require.NoError(t, err)
	require.NotNil(t, d)
	assert.Equal(t, "acme.com", d.Domain)
	businessRepo.AssertExpectations(t)
}

func TestBusinessUseCase_VerifyDomain(t *testing.T) {
	businessRepo := new(testutil.MockBusinessRepo)
	userRepo := new(testutil.MockUserRepo)
	uc := NewBusinessUseCase(businessRepo, userRepo)

	domain := testutil.CreateTestDomain(5, "acme.com", false, false)
	domain.ID = 4
	businessRepo.On("GetUserRole", mock.Anything, int64(5), int64(1)).Return(BusinessRoleAdmin, nil)
	businessRepo.On("GetDomainByVerificationToken", mock.Anything, "verify-tok").Return(domain, nil)
	businessRepo.On("VerifyDomain", mock.Anything, int64(4)).Return(nil)

	d, err := uc.VerifyDomain(context.Background(), 1, 5, "verify-tok")
	require.NoError(t, err)
	require.NotNil(t, d)
	businessRepo.AssertExpectations(t)
}

func TestBusinessUseCase_ToggleDomainAutoJoin(t *testing.T) {
	businessRepo := new(testutil.MockBusinessRepo)
	userRepo := new(testutil.MockUserRepo)
	uc := NewBusinessUseCase(businessRepo, userRepo)

	businessRepo.On("GetUserRole", mock.Anything, int64(5), int64(1)).Return(BusinessRoleAdmin, nil)
	businessRepo.On("UpdateDomainAutoJoin", mock.Anything, int64(4), int64(5), true).Return(nil)

	err := uc.ToggleDomainAutoJoin(context.Background(), 1, 5, 4, true)
	assert.NoError(t, err)
	businessRepo.AssertExpectations(t)
}

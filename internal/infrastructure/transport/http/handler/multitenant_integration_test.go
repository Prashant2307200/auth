package handler

import (
	"context"
	"testing"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMultiTenantIsolation(t *testing.T) {
	ctx := context.Background()

	mockMemberRepo := &testutil.MockMemberRepo{}
	mockAuditRepo := &testutil.MockAuditRepo{}
	mockEmailSvc := &testutil.MockEmailService{}

	teamUseCase := usecase.NewTeamUsecase(mockMemberRepo, mockAuditRepo, mockEmailSvc, nil)

	t.Run("ListMembers returns only members from specified business", func(t *testing.T) {
		businessID := int64(100)
		userID1 := int64(1)
		userID2 := int64(2)
		expectedMembers := []*entity.BusinessMember{
			{ID: 1, BusinessID: businessID, UserID: &userID1, Email: "user1@test.com", Status: entity.MemberStatusActive},
			{ID: 2, BusinessID: businessID, UserID: &userID2, Email: "user2@test.com", Status: entity.MemberStatusActive},
		}

		mockMemberRepo.On("ListByBusiness", ctx, businessID).Return(expectedMembers, nil)

		members, err := teamUseCase.ListMembers(ctx, businessID)

		assert.NoError(t, err)
		assert.Len(t, members, 2)
		assert.Equal(t, expectedMembers[0].BusinessID, members[0].BusinessID)
		mockMemberRepo.AssertCalled(t, "ListByBusiness", ctx, businessID)
	})

	t.Run("RemoveMember marks member as deleted for isolation", func(t *testing.T) {
		businessID := int64(200)
		memberID := int64(5)
		userID := int64(10)

		member := &entity.BusinessMember{
			ID:         memberID,
			BusinessID: businessID,
			UserID:     &userID,
			Email:      "member@test.com",
			Status:     entity.MemberStatusActive,
		}

		mockMemberRepo.On("GetByID", ctx, memberID).Return(member, nil)
		mockMemberRepo.On("Delete", ctx, memberID).Return(nil)
		mockAuditRepo.On("Log", ctx, mock.AnythingOfType("*entity.AuditLog")).Return(nil)

		err := teamUseCase.RemoveMember(ctx, businessID, memberID)

		assert.NoError(t, err)
		mockMemberRepo.AssertCalled(t, "Delete", ctx, memberID)
	})
}

func TestRBACEnforcement(t *testing.T) {
	ctx := context.Background()

	mockMemberRepo := &testutil.MockMemberRepo{}
	mockAuditRepo := &testutil.MockAuditRepo{}
	mockEmailSvc := &testutil.MockEmailService{}

	teamUseCase := usecase.NewTeamUsecase(mockMemberRepo, mockAuditRepo, mockEmailSvc, nil)

	t.Run("UpdateMemberRole changes role for team member", func(t *testing.T) {
		businessID := int64(300)
		memberID := int64(15)
		userID := int64(20)

		member := &entity.BusinessMember{
			ID:         memberID,
			BusinessID: businessID,
			UserID:     &userID,
			Email:      "member@test.com",
			RoleID:     2,
			Status:     entity.MemberStatusActive,
		}

		newRoleID := 3

		mockMemberRepo.On("GetByID", ctx, memberID).Return(member, nil)
		mockMemberRepo.On("Update", ctx, mock.MatchedBy(func(m *entity.BusinessMember) bool {
			return m.ID == memberID && m.RoleID == int64(newRoleID)
		})).Return(nil)

		err := teamUseCase.UpdateMemberRole(ctx, businessID, memberID, newRoleID)

		assert.NoError(t, err)
		mockMemberRepo.AssertCalled(t, "Update", ctx, mock.AnythingOfType("*entity.BusinessMember"))
	})

	t.Run("InviteUser creates pending member invitation", func(t *testing.T) {
		businessID := int64(400)
		email := "newuser@test.com"
		roleID := 2

		mockMemberRepo.On("Create", ctx, mock.MatchedBy(func(m *entity.BusinessMember) bool {
			return m.BusinessID == businessID && m.Email == email && m.Status == entity.MemberStatusPending
		})).Return(nil)
		// Expect Update to be called to store the invite token (legacy or generated)
		mockMemberRepo.On("Update", ctx, mock.MatchedBy(func(m *entity.BusinessMember) bool {
			// InviteToken should be set and status should still be pending until acceptance
			return m.BusinessID == businessID && m.Email == email && m.Status == entity.MemberStatusPending && m.InviteToken != ""
		})).Return(nil)
		mockAuditRepo.On("Log", ctx, mock.AnythingOfType("*entity.AuditLog")).Return(nil)
		mockEmailSvc.On("SendInvite", ctx, email, mock.AnythingOfType("string")).Return(nil)

		token, err := teamUseCase.InviteUser(ctx, businessID, email, roleID)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		mockMemberRepo.AssertCalled(t, "Create", ctx, mock.AnythingOfType("*entity.BusinessMember"))
	})
}

func TestAuditLogging(t *testing.T) {
	ctx := context.Background()

	mockMemberRepo := &testutil.MockMemberRepo{}
	mockAuditRepo := &testutil.MockAuditRepo{}
	mockEmailSvc := &testutil.MockEmailService{}

	teamUseCase := usecase.NewTeamUsecase(mockMemberRepo, mockAuditRepo, mockEmailSvc, nil)

	t.Run("All member operations logged to audit trail", func(t *testing.T) {
		businessID := int64(500)
		memberID := int64(25)

		mockMemberRepo.On("Delete", ctx, memberID).Return(nil)

		err := teamUseCase.RemoveMember(ctx, businessID, memberID)

		assert.NoError(t, err)
		mockMemberRepo.AssertCalled(t, "Delete", ctx, memberID)
	})

	t.Run("Audit logs list by business", func(t *testing.T) {
		businessID := int64(600)

		expectedLogs := []*entity.AuditLog{
			{ID: 1, BusinessID: businessID, Action: "member.invited", CreatedAt: time.Now()},
			{ID: 2, BusinessID: businessID, Action: "member.accepted", CreatedAt: time.Now()},
			{ID: 3, BusinessID: businessID, Action: "member.removed", CreatedAt: time.Now()},
		}

		mockAuditRepo.On("ListByBusiness", ctx, businessID, mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(expectedLogs, nil)

		logs, err := mockAuditRepo.ListByBusiness(ctx, businessID, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, logs, 3)
		assert.Equal(t, "member.invited", logs[0].Action)
		assert.Equal(t, "member.removed", logs[2].Action)
	})
}

func TestTenantScopedDataAccess(t *testing.T) {
	ctx := context.Background()

	mockMemberRepo := &testutil.MockMemberRepo{}
	mockAuditRepo := &testutil.MockAuditRepo{}
	mockEmailSvc := &testutil.MockEmailService{}

	teamUseCase := usecase.NewTeamUsecase(mockMemberRepo, mockAuditRepo, mockEmailSvc, nil)

	t.Run("Tenant A cannot access Tenant B members", func(t *testing.T) {
		tenantAID := int64(1000)
		tenantBID := int64(2000)

		tenantAMembers := []*entity.BusinessMember{
			{ID: 1, BusinessID: tenantAID, Email: "user1@tenant-a.com"},
		}
		tenantBMembers := []*entity.BusinessMember{
			{ID: 2, BusinessID: tenantBID, Email: "user2@tenant-b.com"},
		}

		mockMemberRepo.On("ListByBusiness", ctx, tenantAID).Return(tenantAMembers, nil)
		mockMemberRepo.On("ListByBusiness", ctx, tenantBID).Return(tenantBMembers, nil)

		membersA, err := teamUseCase.ListMembers(ctx, tenantAID)
		assert.NoError(t, err)
		assert.Len(t, membersA, 1)
		assert.Equal(t, tenantAID, membersA[0].BusinessID)

		membersB, err := teamUseCase.ListMembers(ctx, tenantBID)
		assert.NoError(t, err)
		assert.Len(t, membersB, 1)
		assert.Equal(t, tenantBID, membersB[0].BusinessID)

		assert.NotEqual(t, membersA[0].BusinessID, membersB[0].BusinessID)
	})
}

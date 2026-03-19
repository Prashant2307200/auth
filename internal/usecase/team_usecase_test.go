package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	invitetoken "github.com/Prashant2307200/auth-service/pkg/invitetoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTeamUsecase_ListMembers(t *testing.T) {
	memberRepo := new(testutil.MockMemberRepo)
	auditRepo := new(testutil.MockAuditRepo)
	var emailSvc interface {
		SendInvite(context.Context, string, string) error
	}

	members := []*entity.BusinessMember{{ID: 1, BusinessID: 10, Email: "a@x.com", Status: entity.MemberStatusActive}}
	memberRepo.On("ListByBusiness", mock.Anything, int64(10)).Return(members, nil)

	uc := NewTeamUsecase(memberRepo, auditRepo, emailSvc, invitetoken.NewGenerator("test-secret", 24))
	res, err := uc.ListMembers(context.Background(), 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))

	memberRepo.AssertExpectations(t)
}

func TestTeamUsecase_AcceptInvitation_Success(t *testing.T) {
	memberRepo := new(testutil.MockMemberRepo)
	auditRepo := new(testutil.MockAuditRepo)
	var emailSvc interface {
		SendInvite(context.Context, string, string) error
	}

	tokenGen := invitetoken.NewGenerator("test-secret", 24)
	token, _, _ := tokenGen.Generate(5, 10, "newmember@example.com")

	member := &entity.BusinessMember{
		ID:          5,
		BusinessID:  10,
		Email:       "newmember@example.com",
		RoleID:      3,
		Status:      entity.MemberStatusPending,
		InvitedAt:   time.Now().Add(-24 * time.Hour),
		InviteToken: token,
	}

	memberRepo.On("GetByInviteToken", mock.Anything, token).Return(member, nil)
	memberRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *entity.BusinessMember) bool {
		return m.ID == 5 && m.Status == entity.MemberStatusActive && m.AcceptedAt != nil
	})).Return(nil)
	auditRepo.On("Log", mock.Anything, mock.MatchedBy(func(a *entity.AuditLog) bool {
		return a.BusinessID == 10 && a.Action == "accept_invitation"
	})).Return(nil)

	uc := NewTeamUsecase(memberRepo, auditRepo, emailSvc, tokenGen)
	err := uc.AcceptInvitation(context.Background(), token)
	assert.NoError(t, err)

	memberRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestTeamUsecase_AcceptInvitation_InvalidToken(t *testing.T) {
	memberRepo := new(testutil.MockMemberRepo)
	auditRepo := new(testutil.MockAuditRepo)
	var emailSvc interface {
		SendInvite(context.Context, string, string) error
	}

	uc := NewTeamUsecase(memberRepo, auditRepo, emailSvc, invitetoken.NewGenerator("test-secret", 24))
	err := uc.AcceptInvitation(context.Background(), "invalid-token-xyz")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired")
}

func TestTeamUsecase_AcceptInvitation_NotPending(t *testing.T) {
	memberRepo := new(testutil.MockMemberRepo)
	auditRepo := new(testutil.MockAuditRepo)
	var emailSvc interface {
		SendInvite(context.Context, string, string) error
	}

	tokenGen := invitetoken.NewGenerator("test-secret", 24)
	token, _, _ := tokenGen.Generate(5, 10, "existing@example.com")

	member := &entity.BusinessMember{
		ID:          5,
		BusinessID:  10,
		Email:       "existing@example.com",
		Status:      entity.MemberStatusActive,
		AcceptedAt:  &time.Time{},
		InviteToken: token,
	}

	memberRepo.On("GetByInviteToken", mock.Anything, token).Return(member, nil)

	uc := NewTeamUsecase(memberRepo, auditRepo, emailSvc, tokenGen)
	err := uc.AcceptInvitation(context.Background(), token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invitation is not pending")

	memberRepo.AssertExpectations(t)
}

func TestTeamUsecase_AcceptInvitation_MemberNotFound(t *testing.T) {
	memberRepo := new(testutil.MockMemberRepo)
	auditRepo := new(testutil.MockAuditRepo)
	var emailSvc interface {
		SendInvite(context.Context, string, string) error
	}

	tokenGen := invitetoken.NewGenerator("test-secret", 24)
	fakeToken, _, _ := tokenGen.Generate(99, 10, "noone@example.com")
	memberRepo.On("GetByInviteToken", mock.Anything, fakeToken).Return(nil, assert.AnError)

	uc := NewTeamUsecase(memberRepo, auditRepo, emailSvc, tokenGen)
	err := uc.AcceptInvitation(context.Background(), fakeToken)
	assert.Error(t, err)

	memberRepo.AssertExpectations(t)
}

func TestTeamUsecase_AcceptInvitation_WithMemberPrefix(t *testing.T) {
	memberRepo := new(testutil.MockMemberRepo)
	auditRepo := new(testutil.MockAuditRepo)
	var emailSvc interface {
		SendInvite(context.Context, string, string) error
	}

	tokenGen := invitetoken.NewGenerator("test-secret", 24)
	token, _, _ := tokenGen.Generate(7, 10, "newmember@example.com")

	member := &entity.BusinessMember{
		ID:          7,
		BusinessID:  10,
		Email:       "newmember@example.com",
		Status:      entity.MemberStatusPending,
		InvitedAt:   time.Now().Add(-24 * time.Hour),
		InviteToken: token,
	}

	memberRepo.On("GetByInviteToken", mock.Anything, token).Return(member, nil)
	memberRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *entity.BusinessMember) bool {
		return m.ID == 7 && m.Status == entity.MemberStatusActive
	})).Return(nil)
	auditRepo.On("Log", mock.Anything, mock.MatchedBy(func(a *entity.AuditLog) bool {
		return a.BusinessID == 10 && a.Action == "accept_invitation"
	})).Return(nil)

	uc := NewTeamUsecase(memberRepo, auditRepo, emailSvc, tokenGen)
	err := uc.AcceptInvitation(context.Background(), token)
	assert.NoError(t, err)

	memberRepo.AssertExpectations(t)
}

func TestTeamUsecase_RemoveMember_CrossBusinessProtection(t *testing.T) {
	memberRepo := new(testutil.MockMemberRepo)
	auditRepo := new(testutil.MockAuditRepo)
	var emailSvc interface {
		SendInvite(context.Context, string, string) error
	}

	// member belongs to business 10 but we'll attempt to delete under business 99
	member := &entity.BusinessMember{ID: 55, BusinessID: 10, Email: "x@x.com", Status: entity.MemberStatusActive}
	memberRepo.On("GetByID", mock.Anything, int64(55)).Return(member, nil)

	uc := NewTeamUsecase(memberRepo, auditRepo, emailSvc, invitetoken.NewGenerator("test-secret", 24))
	err := uc.RemoveMember(context.Background(), 99, 55)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not belong to business")

	memberRepo.AssertExpectations(t)
}

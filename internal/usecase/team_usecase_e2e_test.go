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

// Full invite -> accept -> list workflow
func TestE2E_InviteAndAcceptWorkflow(t *testing.T) {
	memberRepo := new(testutil.MockMemberRepo)
	auditRepo := new(testutil.MockAuditRepo)
	emailSvc := new(testutil.MockEmailService)

	tokenGen := invitetoken.NewGenerator("e2e-secret", 24)

	// Create will set ID via side-effect in real repo; here we simulate storing
	memberRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		m := args.Get(1).(*entity.BusinessMember)
		m.ID = 42
	})

	// After token generation, Update will be called to persist token
	memberRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	emailSvc.On("SendInvite", mock.Anything, "invitee@example.com", mock.Anything).Return(nil)
	auditRepo.On("Log", mock.Anything, mock.Anything).Return(nil)

	uc := NewTeamUsecase(memberRepo, auditRepo, emailSvc, tokenGen)
	token, err := uc.InviteUser(context.Background(), 100, "invitee@example.com", 2)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Prepare member lookup on accept
	member := &entity.BusinessMember{ID: 42, BusinessID: 100, Email: "invitee@example.com", Status: entity.MemberStatusPending, InviteToken: token, InvitedAt: time.Now()}
	memberRepo.On("GetByInviteToken", mock.Anything, token).Return(member, nil)
	memberRepo.On("Update", mock.Anything, mock.MatchedBy(func(m *entity.BusinessMember) bool { return m.Status == entity.MemberStatusActive })).Return(nil)

	err = uc.AcceptInvitation(context.Background(), token)
	assert.NoError(t, err)

	// List should return active member
	memberRepo.On("ListByBusiness", mock.Anything, int64(100)).Return([]*entity.BusinessMember{member}, nil)
	list, err := uc.ListMembers(context.Background(), 100)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(list))

	memberRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
	emailSvc.AssertExpectations(t)
}

// Expired token cannot be accepted
func TestE2E_ExpiredInviteCannotBeAccepted(t *testing.T) {
	memberRepo := new(testutil.MockMemberRepo)
	auditRepo := new(testutil.MockAuditRepo)
	var emailSvc *testutil.MockEmailService

	tokenGen := invitetoken.NewGenerator("e2e-secret", -1) // expired immediately
	token, _, _ := tokenGen.Generate(7, 200, "old@example.com")

	member := &entity.BusinessMember{ID: 7, BusinessID: 200, Email: "old@example.com", Status: entity.MemberStatusPending, InviteToken: token, InvitedAt: time.Now().Add(-48 * time.Hour)}
	memberRepo.On("GetByInviteToken", mock.Anything, token).Return(member, nil)

	uc := NewTeamUsecase(memberRepo, auditRepo, emailSvc, tokenGen)
	err := uc.AcceptInvitation(context.Background(), token)
	assert.Error(t, err)
}

// Multiple users in different businesses remain isolated
func TestE2E_MultipleUsersMultipleBusinesses(t *testing.T) {
	memberRepo := new(testutil.MockMemberRepo)
	auditRepo := new(testutil.MockAuditRepo)
	var emailSvc interface {
		SendInvite(context.Context, string, string) error
	}
	tokenGen := invitetoken.NewGenerator("e2e-secret", 24)

	// two members created in different businesses
	member1 := &entity.BusinessMember{ID: 11, BusinessID: 1, Email: "a@x.com", Status: entity.MemberStatusActive}
	member2 := &entity.BusinessMember{ID: 12, BusinessID: 2, Email: "b@x.com", Status: entity.MemberStatusActive}

	memberRepo.On("ListByBusiness", mock.Anything, int64(1)).Return([]*entity.BusinessMember{member1}, nil)
	memberRepo.On("ListByBusiness", mock.Anything, int64(2)).Return([]*entity.BusinessMember{member2}, nil)

	uc := NewTeamUsecase(memberRepo, auditRepo, emailSvc, tokenGen)
	l1, err := uc.ListMembers(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(l1))
	l2, err := uc.ListMembers(context.Background(), 2)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(l2))

	memberRepo.AssertExpectations(t)
}

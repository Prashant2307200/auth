package usecase

import (
	"context"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	invitetoken "github.com/Prashant2307200/auth-service/pkg/invitetoken"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRevokeInvitation_SuccessfullyRevokes(t *testing.T) {
	memberRepo := &testutil.MockMemberRepo{}
	auditRepo := &testutil.MockAuditRepo{}
	tokenGen := &invitetoken.Generator{}

	member := &entity.BusinessMember{
		ID:          1,
		BusinessID:  100,
		Email:       "user@example.com",
		Status:      entity.MemberStatusPending,
		InviteToken: "valid-token",
	}

	memberRepo.On("GetByInviteToken", mock.Anything, "valid-token").Return(member, nil)

	revokedMember := *member
	revokedMember.Status = entity.MemberStatusRevoked
	revokedMember.InviteToken = ""
	memberRepo.On("Update", mock.Anything, &revokedMember).Return(nil)

	auditRepo.On("Log", mock.Anything, mock.MatchedBy(func(al *entity.AuditLog) bool {
		return al.Action == "revoke_invitation" && al.BusinessID == 100
	})).Return(nil)

	tu := &teamUsecase{memberRepo: memberRepo, auditRepo: auditRepo, tokenGen: tokenGen}

	err := tu.RevokeInvitation(context.Background(), "valid-token")
	assert.NoError(t, err)
	memberRepo.AssertExpectations(t)
	auditRepo.AssertExpectations(t)
}

func TestRevokeInvitation_TokenNotFound(t *testing.T) {
	memberRepo := &testutil.MockMemberRepo{}
	memberRepo.On("GetByInviteToken", mock.Anything, "nonexistent").Return(nil, testutil.ErrNotFound)

	tu := &teamUsecase{memberRepo: memberRepo}

	err := tu.RevokeInvitation(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRevokeInvitation_AlreadyActive(t *testing.T) {
	memberRepo := &testutil.MockMemberRepo{}

	member := &entity.BusinessMember{
		ID:          1,
		BusinessID:  100,
		Status:      entity.MemberStatusActive,
		InviteToken: "already-accepted",
	}
	memberRepo.On("GetByInviteToken", mock.Anything, "already-accepted").Return(member, nil)

	tu := &teamUsecase{memberRepo: memberRepo}

	err := tu.RevokeInvitation(context.Background(), "already-accepted")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot revoke")
}

func TestRevokeInvitation_AlreadyRevoked(t *testing.T) {
	memberRepo := &testutil.MockMemberRepo{}

	member := &entity.BusinessMember{
		ID:          1,
		BusinessID:  100,
		Status:      entity.MemberStatusRevoked,
		InviteToken: "already-revoked",
	}
	memberRepo.On("GetByInviteToken", mock.Anything, "already-revoked").Return(member, nil)

	tu := &teamUsecase{memberRepo: memberRepo}

	err := tu.RevokeInvitation(context.Background(), "already-revoked")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot revoke")
}

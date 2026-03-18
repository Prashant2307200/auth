package usecase

import (
	"context"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
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

	uc := NewTeamUsecase(memberRepo, auditRepo, emailSvc)
	res, err := uc.ListMembers(context.Background(), 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))

	memberRepo.AssertExpectations(t)
}

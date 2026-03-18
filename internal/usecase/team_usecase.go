package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/repository"
)

type TeamUsecase interface {
	InviteUser(ctx context.Context, businessID int64, email string, role int) error
	AcceptInvitation(ctx context.Context, inviteToken string) error
	ListMembers(ctx context.Context, businessID int64) ([]*entity.BusinessMember, error)
	RemoveMember(ctx context.Context, businessID int64, memberID int64) error
	UpdateMemberRole(ctx context.Context, businessID int64, memberID int64, newRole int) error
}

type EmailService interface {
	SendInvite(ctx context.Context, to string, token string) error
}

type teamUsecase struct {
	memberRepo repository.MemberRepository
	auditRepo  repository.AuditRepository
	emailSvc   EmailService
}

func NewTeamUsecase(m repository.MemberRepository, a repository.AuditRepository, e EmailService) TeamUsecase {
	return &teamUsecase{memberRepo: m, auditRepo: a, emailSvc: e}
}

var ErrNotImplemented = errors.New("not implemented")

func (t *teamUsecase) InviteUser(ctx context.Context, businessID int64, email string, role int) error {
	if t.memberRepo == nil {
		return ErrNotImplemented
	}
	bm := &entity.BusinessMember{
		BusinessID: businessID,
		Email:      email,
		RoleID:     int64(role),
		Status:     entity.MemberStatusPending,
		InvitedAt:  time.Now(),
	}
	if err := t.memberRepo.Create(ctx, bm); err != nil {
		return err
	}
	if t.auditRepo != nil {
		_ = t.auditRepo.Log(ctx, &entity.AuditLog{BusinessID: businessID, Action: "invite_user", UserID: 0, CreatedAt: time.Now()})
	}
	// Email sending intentionally omitted (mock in tests)
	return nil
}

func (t *teamUsecase) AcceptInvitation(ctx context.Context, inviteToken string) error {
	// Token-based invites not implemented in this minimal version
	return ErrNotImplemented
}

func (t *teamUsecase) ListMembers(ctx context.Context, businessID int64) ([]*entity.BusinessMember, error) {
	if t.memberRepo == nil {
		return nil, ErrNotImplemented
	}
	return t.memberRepo.ListByBusiness(ctx, businessID)
}

func (t *teamUsecase) RemoveMember(ctx context.Context, businessID int64, memberID int64) error {
	if t.memberRepo == nil {
		return ErrNotImplemented
	}
	// Soft delete semantics depend on repo; call Delete for now
	return t.memberRepo.Delete(ctx, memberID)
}

func (t *teamUsecase) UpdateMemberRole(ctx context.Context, businessID int64, memberID int64, newRole int) error {
	if t.memberRepo == nil {
		return ErrNotImplemented
	}
	m, err := t.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return err
	}
	if m.BusinessID != businessID {
		return errors.New("member does not belong to business")
	}
	m.RoleID = int64(newRole)
	return t.memberRepo.Update(ctx, m)
}

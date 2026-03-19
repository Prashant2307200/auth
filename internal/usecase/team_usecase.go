package usecase

import (
	"context"
	"errors"
	"fmt"
	"strconv"
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
	if t.memberRepo == nil {
		return ErrNotImplemented
	}

	var memberID int64
	if _, err := fmt.Sscanf(inviteToken, "member_%d", &memberID); err != nil {
		if id, parseErr := strconv.ParseInt(inviteToken, 10, 64); parseErr == nil {
			memberID = id
		} else {
			return errors.New("invalid invite token format")
		}
	}

	member, err := t.memberRepo.GetByID(ctx, memberID)
	if err != nil {
		return err
	}

	if member.Status != entity.MemberStatusPending {
		return errors.New("invitation is not pending")
	}

	now := time.Now()
	member.Status = entity.MemberStatusActive
	member.AcceptedAt = &now

	if err := t.memberRepo.Update(ctx, member); err != nil {
		return err
	}

	if t.auditRepo != nil {
		_ = t.auditRepo.Log(ctx, &entity.AuditLog{
			BusinessID: member.BusinessID,
			Action:     "accept_invitation",
			UserID:     0,
			CreatedAt:  time.Now(),
		})
	}

	return nil
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

package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/repository"
	invitetoken "github.com/Prashant2307200/auth-service/pkg/invitetoken"
)

type TeamUsecase interface {
	InviteUser(ctx context.Context, businessID int64, email string, role int) (string, error)
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
	tokenGen   *invitetoken.Generator
}

func NewTeamUsecase(m repository.MemberRepository, a repository.AuditRepository, e EmailService, tg *invitetoken.Generator) TeamUsecase {
	return &teamUsecase{memberRepo: m, auditRepo: a, emailSvc: e, tokenGen: tg}
}

var ErrNotImplemented = errors.New("not implemented")

func (t *teamUsecase) InviteUser(ctx context.Context, businessID int64, email string, role int) (string, error) {
	if t.memberRepo == nil {
		return "", ErrNotImplemented
	}
	bm := &entity.BusinessMember{
		BusinessID: businessID,
		Email:      email,
		RoleID:     int64(role),
		Status:     entity.MemberStatusPending,
		InvitedAt:  time.Now(),
	}
	if err := t.memberRepo.Create(ctx, bm); err != nil {
		return "", err
	}
	var token string
	var expiresAt time.Time
	if t.tokenGen != nil {
		var err error
		token, expiresAt, err = t.tokenGen.Generate(bm.ID, businessID, email)
		if err != nil {
			return "", fmt.Errorf("failed to generate invite token: %w", err)
		}
		bm.InviteToken = token
		bm.TokenExpiresAt = &expiresAt
		if err := t.memberRepo.Update(ctx, bm); err != nil {
			return "", fmt.Errorf("failed to store invite token: %w", err)
		}
	} else {
		// fallback legacy token for environments without token generator
		token = fmt.Sprintf("member_%d", bm.ID)
		bm.InviteToken = token
		if err := t.memberRepo.Update(ctx, bm); err != nil {
			return "", fmt.Errorf("failed to store invite token: %w", err)
		}
	}
	// send invite email if service configured
	if t.emailSvc != nil {
		_ = t.emailSvc.SendInvite(ctx, email, token)
	}
	if t.auditRepo != nil {
		_ = t.auditRepo.Log(ctx, &entity.AuditLog{BusinessID: businessID, Action: "invite_user", UserID: 0, CreatedAt: time.Now()})
	}
	return token, nil
}

func (t *teamUsecase) AcceptInvitation(ctx context.Context, inviteToken string) error {
	if t.memberRepo == nil {
		return ErrNotImplemented
	}

	var claims *invitetoken.InviteTokenClaims
	if t.tokenGen != nil {
		var err error
		claims, err = t.tokenGen.Validate(inviteToken)
		if err != nil {
			return fmt.Errorf("invalid or expired invite token: %w", err)
		}
	}

	member, err := t.memberRepo.GetByInviteToken(ctx, inviteToken)
	if err != nil {
		return fmt.Errorf("invite not found: %w", err)
	}

	if member.Status != entity.MemberStatusPending {
		return errors.New("invitation is not pending")
	}

	if claims != nil && (claims.MemberID != member.ID || claims.Email != member.Email) {
		return errors.New("token does not match member")
	}

	now := time.Now()
	member.Status = entity.MemberStatusActive
	member.AcceptedAt = &now
	member.InviteToken = ""

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

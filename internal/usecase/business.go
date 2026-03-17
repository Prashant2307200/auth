package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
)

const (
	BusinessRoleMember = 0
	BusinessRoleAdmin  = 1
	BusinessRoleOwner  = 2
)

type BusinessUseCase struct {
	BusinessRepo interfaces.BusinessRepo
	UserRepo     interfaces.UserRepo
}

func NewBusinessUseCase(businessRepo interfaces.BusinessRepo, userRepo interfaces.UserRepo) *BusinessUseCase {
	return &BusinessUseCase{
		BusinessRepo: businessRepo,
		UserRepo:     userRepo,
	}
}

func (uc *BusinessUseCase) CreateBusiness(ctx context.Context, creatorID int64, business *entity.Business) (*entity.Business, error) {
	if business == nil {
		return nil, fmt.Errorf("business cannot be nil")
	}

	existing, _ := uc.BusinessRepo.GetBySlug(ctx, business.Slug)
	if existing != nil {
		return nil, fmt.Errorf("business with slug %s already exists", business.Slug)
	}

	business.OwnerID = creatorID
	id, err := uc.BusinessRepo.Create(ctx, business)
	if err != nil {
		return nil, err
	}
	if err := uc.BusinessRepo.AddUser(ctx, id, creatorID, BusinessRoleOwner); err != nil {
		return nil, err
	}
	return uc.BusinessRepo.GetById(ctx, id)
}

func (uc *BusinessUseCase) GetBusinessByID(ctx context.Context, id int64) (*entity.Business, error) {
	return uc.BusinessRepo.GetById(ctx, id)
}

func (uc *BusinessUseCase) UpdateBusiness(ctx context.Context, requesterID int64, businessID int64, business *entity.Business) error {
	role, err := uc.BusinessRepo.GetUserRole(ctx, businessID, requesterID)
	if err != nil {
		return fmt.Errorf("not allowed to update business")
	}
	if role < BusinessRoleAdmin {
		return fmt.Errorf("not allowed to update business")
	}
	return uc.BusinessRepo.Update(ctx, businessID, business)
}

func (uc *BusinessUseCase) DeleteBusiness(ctx context.Context, requesterID int64, businessID int64) error {
	role, err := uc.BusinessRepo.GetUserRole(ctx, businessID, requesterID)
	if err != nil {
		return fmt.Errorf("not allowed to delete business")
	}
	if role != BusinessRoleOwner {
		return fmt.Errorf("only owner can delete business")
	}
	return uc.BusinessRepo.Delete(ctx, businessID)
}

func (uc *BusinessUseCase) AddUserToBusiness(ctx context.Context, requesterID int64, businessID int64, userID int64, role int) error {
	if role < BusinessRoleMember || role > BusinessRoleOwner {
		return fmt.Errorf("invalid role")
	}
	requesterRole, err := uc.BusinessRepo.GetUserRole(ctx, businessID, requesterID)
	if err != nil {
		return fmt.Errorf("not allowed to add user to business")
	}
	if requesterRole < BusinessRoleAdmin {
		return fmt.Errorf("not allowed to add user to business")
	}
	if _, err := uc.UserRepo.GetById(ctx, userID); err != nil {
		return fmt.Errorf("failed to validate user: %w", err)
	}
	return uc.BusinessRepo.AddUser(ctx, businessID, userID, role)
}

func (uc *BusinessUseCase) RemoveUserFromBusiness(ctx context.Context, requesterID int64, businessID int64, userID int64) error {
	requesterRole, err := uc.BusinessRepo.GetUserRole(ctx, businessID, requesterID)
	if err != nil {
		return fmt.Errorf("not allowed to remove user from business")
	}
	if requesterRole < BusinessRoleAdmin {
		return fmt.Errorf("not allowed to remove user from business")
	}

	targetRole, err := uc.BusinessRepo.GetUserRole(ctx, businessID, userID)
	if err != nil {
		return fmt.Errorf("target user is not part of business")
	}
	if targetRole == BusinessRoleOwner {
		return fmt.Errorf("owner cannot be removed from business")
	}
	return uc.BusinessRepo.RemoveUser(ctx, businessID, userID)
}

func (uc *BusinessUseCase) GetBusinessUsers(ctx context.Context, requesterID int64, businessID int64) ([]*entity.User, error) {
	if _, err := uc.BusinessRepo.GetUserRole(ctx, businessID, requesterID); err != nil {
		return nil, fmt.Errorf("not allowed to access business users")
	}
	return uc.BusinessRepo.GetUsers(ctx, businessID)
}

func (uc *BusinessUseCase) GetUserBusinesses(ctx context.Context, userID int64) ([]*entity.Business, error) {
	return uc.BusinessRepo.GetUserBusinesses(ctx, userID)
}

func (uc *BusinessUseCase) CreateInvite(ctx context.Context, requesterID int64, businessID int64, email string, role int) (*entity.BusinessInvite, error) {
	requesterRole, err := uc.BusinessRepo.GetUserRole(ctx, businessID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("not allowed to create invite")
	}
	if requesterRole < BusinessRoleAdmin {
		return nil, fmt.Errorf("not allowed to create invite")
	}
	if role < BusinessRoleMember || role > BusinessRoleOwner {
		return nil, fmt.Errorf("invalid role")
	}
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invite token: %w", err)
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	invite := &entity.BusinessInvite{
		BusinessID: businessID,
		Email:      email,
		Role:       role,
		InvitedBy:  requesterID,
		Token:      token,
		ExpiresAt:  expiresAt,
		Status:     entity.InviteStatusPending,
	}
	id, err := uc.BusinessRepo.CreateInvite(ctx, invite)
	if err != nil {
		return nil, err
	}
	invite.ID = id
	return invite, nil
}

func (uc *BusinessUseCase) ListInvites(ctx context.Context, requesterID int64, businessID int64) ([]*entity.BusinessInvite, error) {
	role, err := uc.BusinessRepo.GetUserRole(ctx, businessID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("not allowed to list invites")
	}
	if role < BusinessRoleAdmin {
		return nil, fmt.Errorf("not allowed to list invites")
	}
	return uc.BusinessRepo.ListInvites(ctx, businessID)
}

func (uc *BusinessUseCase) RevokeInvite(ctx context.Context, requesterID int64, businessID int64, inviteID int64) error {
	requesterRole, err := uc.BusinessRepo.GetUserRole(ctx, businessID, requesterID)
	if err != nil {
		return fmt.Errorf("not allowed to revoke invite")
	}
	if requesterRole < BusinessRoleAdmin {
		return fmt.Errorf("not allowed to revoke invite")
	}
	return uc.BusinessRepo.RevokeInvite(ctx, inviteID, businessID)
}

func (uc *BusinessUseCase) AddDomain(ctx context.Context, requesterID int64, businessID int64, domain string) (*entity.BusinessDomain, error) {
	requesterRole, err := uc.BusinessRepo.GetUserRole(ctx, businessID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("not allowed to add domain")
	}
	if requesterRole < BusinessRoleAdmin {
		return nil, fmt.Errorf("not allowed to add domain")
	}
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		return nil, fmt.Errorf("domain cannot be empty")
	}
	existing, _ := uc.BusinessRepo.GetDomain(ctx, businessID, domain)
	if existing != nil {
		return nil, fmt.Errorf("domain already added")
	}
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %w", err)
	}
	d := &entity.BusinessDomain{
		BusinessID:        businessID,
		Domain:            domain,
		Verified:          false,
		AutoJoinEnabled:   false,
		VerificationToken:  token,
	}
	id, err := uc.BusinessRepo.CreateDomain(ctx, d)
	if err != nil {
		return nil, err
	}
	d.ID = id
	return d, nil
}

func (uc *BusinessUseCase) VerifyDomain(ctx context.Context, requesterID int64, businessID int64, verificationToken string) (*entity.BusinessDomain, error) {
	requesterRole, err := uc.BusinessRepo.GetUserRole(ctx, businessID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("not allowed to verify domain")
	}
	if requesterRole < BusinessRoleAdmin {
		return nil, fmt.Errorf("not allowed to verify domain")
	}
	d, err := uc.BusinessRepo.GetDomainByVerificationToken(ctx, verificationToken)
	if err != nil {
		return nil, fmt.Errorf("invalid verification token")
	}
	if d.BusinessID != businessID {
		return nil, fmt.Errorf("domain does not belong to this business")
	}
	if err := uc.BusinessRepo.VerifyDomain(ctx, d.ID); err != nil {
		return nil, err
	}
	d.Verified = true
	now := time.Now()
	d.VerifiedAt = &now
	return d, nil
}

func (uc *BusinessUseCase) ToggleDomainAutoJoin(ctx context.Context, requesterID int64, businessID int64, domainID int64, enabled bool) error {
	requesterRole, err := uc.BusinessRepo.GetUserRole(ctx, businessID, requesterID)
	if err != nil {
		return fmt.Errorf("not allowed to update domain")
	}
	if requesterRole < BusinessRoleAdmin {
		return fmt.Errorf("not allowed to update domain")
	}
	return uc.BusinessRepo.UpdateDomainAutoJoin(ctx, domainID, businessID, enabled)
}

func generateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

package seeder

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
	"github.com/Prashant2307200/auth-service/pkg/db"
	"github.com/Prashant2307200/auth-service/pkg/hash"
)

func isNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || errors.Is(err, db.ErrNotFound)
}

func SeedUsers(ctx context.Context, userRepo interfaces.UserRepo) error {
	for _, seedUser := range SeedUsersData {
		existingUser, err := userRepo.GetByEmail(ctx, seedUser.Email)
		if err != nil && err != sql.ErrNoRows {
			slog.Error("Error checking existing user", slog.String("email", seedUser.Email), slog.Any("error", err))
			return err
		}

		if existingUser != nil {
			slog.Info("User already exists, skipping", slog.String("email", seedUser.Email))
			continue
		}

		hashedPassword, err := hash.HashPassword(seedUser.Password)
		if err != nil {
			slog.Error("Error hashing password", slog.String("email", seedUser.Email), slog.Any("error", err))
			return err
		}

		user := &entity.User{
			Username: seedUser.Username,
			Email:    seedUser.Email,
			Password: hashedPassword,
			Role:     seedUser.Role,
		}

		id, err := userRepo.Create(ctx, user)
		if err != nil {
			slog.Error("Error creating user", slog.String("email", seedUser.Email), slog.Any("error", err))
			return err
		}

		slog.Info("Seeded user", slog.Int64("id", id), slog.String("email", seedUser.Email), slog.String("username", seedUser.Username))
	}

	return nil
}

func SeedBusinesses(ctx context.Context, businessRepo interfaces.BusinessRepo, userRepo interfaces.UserRepo) error {
	for _, sb := range SeedBusinessesData {
		_, err := businessRepo.GetBySlug(ctx, sb.Slug)
		if err != nil && !isNotFound(err) {
			slog.Error("Error checking existing business", slog.String("slug", sb.Slug), slog.Any("error", err))
			return err
		}
		if err == nil {
			slog.Info("Business already exists, skipping", slog.String("slug", sb.Slug))
			continue
		}

		owner, err := userRepo.GetByEmail(ctx, sb.OwnerEmail)
		if err != nil || owner == nil {
			slog.Error("Owner user not found for business", slog.String("owner_email", sb.OwnerEmail), slog.String("slug", sb.Slug))
			if err != nil {
				return err
			}
			return errors.New("owner user not found: " + sb.OwnerEmail)
		}

		signupPolicy := sb.SignupPolicy
		if signupPolicy == "" {
			signupPolicy = entity.SignupPolicyClosed
		}

		business := &entity.Business{
			Name:         sb.Name,
			Slug:         sb.Slug,
			Email:        sb.Email,
			OwnerID:      owner.ID,
			SignupPolicy: signupPolicy,
		}

		id, err := businessRepo.CreateWithOwner(ctx, business, owner.ID)
		if err != nil {
			slog.Error("Error creating business", slog.String("slug", sb.Slug), slog.Any("error", err))
			return err
		}

		slog.Info("Seeded business", slog.Int64("id", id), slog.String("slug", sb.Slug))
	}

	return nil
}

func SeedInvites(ctx context.Context, businessRepo interfaces.BusinessRepo, userRepo interfaces.UserRepo) error {
	for _, si := range SeedInvitesData {
		_, err := businessRepo.GetInviteByToken(ctx, si.Token)
		if err != nil && !isNotFound(err) {
			slog.Error("Error checking existing invite", slog.String("token", si.Token), slog.Any("error", err))
			return err
		}
		if err == nil {
			slog.Info("Invite already exists, skipping", slog.String("token", si.Token))
			continue
		}

		biz, err := businessRepo.GetBySlug(ctx, si.BusinessSlug)
		if err != nil || biz == nil {
			slog.Error("Business not found for invite", slog.String("slug", si.BusinessSlug))
			if err != nil {
				return err
			}
			return errors.New("business not found: " + si.BusinessSlug)
		}

		invitedBy, err := userRepo.GetByEmail(ctx, si.InvitedByEmail)
		if err != nil || invitedBy == nil {
			slog.Error("Inviter user not found", slog.String("email", si.InvitedByEmail))
			if err != nil {
				return err
			}
			return errors.New("inviter user not found: " + si.InvitedByEmail)
		}

		invite := &entity.BusinessInvite{
			BusinessID: biz.ID,
			Email:      si.Email,
			Role:       si.Role,
			InvitedBy:  invitedBy.ID,
			Token:      si.Token,
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
			Status:     entity.InviteStatusPending,
		}

		id, err := businessRepo.CreateInvite(ctx, invite)
		if err != nil {
			slog.Error("Error creating invite", slog.String("email", si.Email), slog.Any("error", err))
			return err
		}

		slog.Info("Seeded invite", slog.Int64("id", id), slog.String("email", si.Email))
	}

	return nil
}

func SeedDomains(ctx context.Context, businessRepo interfaces.BusinessRepo) error {
	for _, sd := range SeedDomainsData {
		biz, err := businessRepo.GetBySlug(ctx, sd.BusinessSlug)
		if err != nil || biz == nil {
			slog.Error("Business not found for domain", slog.String("slug", sd.BusinessSlug))
			if err != nil {
				return err
			}
			return errors.New("business not found: " + sd.BusinessSlug)
		}

		existing, err := businessRepo.GetDomain(ctx, biz.ID, sd.Domain)
		if err != nil && !isNotFound(err) {
			slog.Error("Error checking existing domain", slog.String("domain", sd.Domain), slog.Any("error", err))
			return err
		}
		if existing != nil {
			slog.Info("Domain already exists, skipping", slog.String("domain", sd.Domain))
			continue
		}

		domain := &entity.BusinessDomain{
			BusinessID:      biz.ID,
			Domain:          sd.Domain,
			Verified:        sd.Verified,
			AutoJoinEnabled: sd.AutoJoinEnabled,
		}
		if sd.Verified {
			now := time.Now()
			domain.VerifiedAt = &now
		}

		id, err := businessRepo.CreateDomain(ctx, domain)
		if err != nil {
			slog.Error("Error creating domain", slog.String("domain", sd.Domain), slog.Any("error", err))
			return err
		}

		slog.Info("Seeded domain", slog.Int64("id", id), slog.String("domain", sd.Domain))
	}

	return nil
}

func SeedAll(ctx context.Context, userRepo interfaces.UserRepo, businessRepo interfaces.BusinessRepo) error {
	if err := SeedUsers(ctx, userRepo); err != nil {
		return err
	}
	if err := SeedBusinesses(ctx, businessRepo, userRepo); err != nil {
		return err
	}
	if err := SeedInvites(ctx, businessRepo, userRepo); err != nil {
		return err
	}
	if err := SeedDomains(ctx, businessRepo); err != nil {
		return err
	}
	return nil
}

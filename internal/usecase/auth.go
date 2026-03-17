package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
	"github.com/Prashant2307200/auth-service/pkg/hash"
)

// RegisterOptions carries optional onboarding params for registration.
type RegisterOptions struct {
	InviteToken  string
	BusinessSlug string
}

type AuthUseCase struct {
	UserRepo     interfaces.UserRepo
	BusinessRepo interfaces.BusinessRepo
	TokenService interfaces.TokenService
	CloudService interfaces.CloudService
}

func NewAuthUseCase(r interfaces.UserRepo, br interfaces.BusinessRepo, s interfaces.TokenService, c interfaces.CloudService) *AuthUseCase {
	return &AuthUseCase{
		UserRepo:     r,
		BusinessRepo: br,
		TokenService: s,
		CloudService: c,
	}
}

func (uc *AuthUseCase) RegisterUser(ctx context.Context, user *entity.User, opts *RegisterOptions) (string, string, error) {
	existingUser, err := uc.UserRepo.GetByEmail(ctx, user.Email)
	if err != nil && err != sql.ErrNoRows {
		slog.Error("Failed to check existing user", slog.String("email", user.Email), slog.Any("error", err))
		return "", "", fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return "", "", fmt.Errorf("user with email %s already exists", user.Email)
	}

	user.Password, err = hash.HashPassword(user.Password)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash password: %w", err)
	}

	id, err := uc.UserRepo.Create(ctx, user)
	if err != nil {
		return "", "", fmt.Errorf("failed to create user: %w", err)
	}

	if err := uc.applyOnboarding(ctx, id, user, opts); err != nil {
		slog.Error("Onboarding failed after user creation", slog.Int64("user_id", id), slog.Any("error", err))
		return "", "", err
	}

	refreshToken, err := uc.TokenService.GenerateRefreshToken(id)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	err = uc.TokenService.StoreRefreshToken(ctx, id, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	accessToken, err := uc.TokenService.GenerateAccessToken(id)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (uc *AuthUseCase) applyOnboarding(ctx context.Context, userID int64, user *entity.User, opts *RegisterOptions) error {
	if uc.BusinessRepo == nil {
		return nil
	}
	if opts != nil && opts.InviteToken != "" {
		inv, err := uc.BusinessRepo.GetInviteByToken(ctx, opts.InviteToken)
		if err != nil {
			return fmt.Errorf("invalid invite token: %w", err)
		}
		if inv.Email != user.Email {
			return fmt.Errorf("invite email does not match registration email")
		}
		if inv.Status != entity.InviteStatusPending {
			return fmt.Errorf("invite already used or revoked")
		}
		if inv.ExpiresAt.Before(time.Now()) {
			return fmt.Errorf("invite expired")
		}
		if err := uc.BusinessRepo.AddUserIfNotExists(ctx, inv.BusinessID, userID, inv.Role); err != nil {
			return fmt.Errorf("failed to add user to business: %w", err)
		}
		return uc.BusinessRepo.AcceptInvite(ctx, inv.ID)
	}
	if opts != nil && opts.BusinessSlug != "" {
		biz, err := uc.BusinessRepo.GetBySlug(ctx, opts.BusinessSlug)
		if err != nil {
			return fmt.Errorf("business not found: %w", err)
		}
		if biz.SignupPolicy != entity.SignupPolicyOpen {
			return fmt.Errorf("business does not allow open signup")
		}
		return uc.BusinessRepo.AddUserIfNotExists(ctx, biz.ID, userID, 0)
	}
	parts := strings.Split(user.Email, "@")
	if len(parts) != 2 {
		return nil
	}
	domain := strings.ToLower(parts[1])
	biz, err := uc.BusinessRepo.FindAutoJoinBusinessByEmailDomain(ctx, domain)
	if err != nil || biz == nil {
		return nil
	}
	return uc.BusinessRepo.AddUserIfNotExists(ctx, biz.ID, userID, 0)
}

func (uc *AuthUseCase) LoginUser(ctx context.Context, email string, password string) (string, string, error) {

	existingUser, err := uc.UserRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", fmt.Errorf("user not found: %w", err)
		}
		return "", "", err
	}

	err = hash.CheckPassword(existingUser.Password, password)
	if err != nil {
		return "", "", fmt.Errorf("invalid password: %w", err)
	}

	refreshToken, err := uc.TokenService.GenerateRefreshToken(existingUser.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	err = uc.TokenService.StoreRefreshToken(ctx, existingUser.ID, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	accessToken, err := uc.TokenService.GenerateAccessToken(existingUser.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (uc *AuthUseCase) LogoutUser(ctx context.Context, userID int64) error {

	err := uc.TokenService.RemoveRefreshToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to remove refresh token: %w", err)
	}

	return nil
}

func (uc *AuthUseCase) GetAuthUserProfile(ctx context.Context, authUserId int64) (*entity.User, error) {

	user, err := uc.UserRepo.GetById(ctx, authUserId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return user, nil
}

func (uc *AuthUseCase) UpdateAuthUserProfile(ctx context.Context, authUserId int64, user *entity.User) error {

	err := uc.UserRepo.UpdateById(ctx, authUserId, user)
	if err != nil {
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (uc *AuthUseCase) RefreshSession(ctx context.Context, refreshToken string) (string, string, error) {

	userID, err := uc.TokenService.VerifyRefreshToken(ctx, refreshToken)
	if err != nil {
		slog.Error("Failed to verify refresh token", slog.Any("error", err))
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	parsedUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		slog.Error("Failed to parse user ID from token", slog.String("user_id", userID), slog.Any("error", err))
		return "", "", fmt.Errorf("invalid user ID in token: %w", err)
	}

	storedToken, err := uc.TokenService.GetRefreshToken(ctx, parsedUserID)
	if err != nil {
		slog.Error("Failed to get stored refresh token", slog.Int64("user_id", parsedUserID), slog.Any("error", err))
		return "", "", fmt.Errorf("refresh token not found in storage: %w", err)
	}

	if storedToken != refreshToken {
		slog.Error("Refresh token mismatch", slog.Int64("user_id", parsedUserID))
		return "", "", errors.New("refresh token does not match stored token")
	}

	newRefreshToken, err := uc.TokenService.GenerateRefreshToken(parsedUserID)
	if err != nil {
		slog.Error("Failed to generate new refresh token", slog.Int64("user_id", parsedUserID), slog.Any("error", err))
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	newAccessToken, err := uc.TokenService.GenerateAccessToken(parsedUserID)
	if err != nil {
		slog.Error("Failed to generate new access token", slog.Int64("user_id", parsedUserID), slog.Any("error", err))
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	err = uc.TokenService.StoreRefreshToken(ctx, parsedUserID, newRefreshToken)
	if err != nil {
		slog.Error("Failed to store new refresh token", slog.Int64("user_id", parsedUserID), slog.Any("error", err))
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return newRefreshToken, newAccessToken, nil
}

func (uc *AuthUseCase) GetPublicKey() ([]byte, error) {

	pubKey, err := uc.TokenService.GetPublicKeyPEM()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	return pubKey, nil
}
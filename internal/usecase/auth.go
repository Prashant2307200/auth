package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/prashant2307200/auth/pkg/hash"
)

type AuthUseCase struct {
	UserRepo     interfaces.UserRepo
	TokenService interfaces.TokenService
	CloudService interfaces.CloudService
}

func NewAuthUseCase(r interfaces.UserRepo, s interfaces.TokenService, c interfaces.CloudService) *AuthUseCase {
	return &AuthUseCase{
		UserRepo:     r,
		TokenService: s,
		CloudService: c,
	}
}

func (uc *AuthUseCase) RegisterUser(ctx context.Context, user *entity.User) (string, string, error) {

	_, err := uc.UserRepo.GetByEmail(ctx, user.Email)
	if err != nil && err != sql.ErrNoRows {
		slog.Error("Check user error", slog.Any("err", err))
		return "", "", fmt.Errorf("failed to check existing user: %w", err)
	}

	user.Password, err = hash.HashPassword(user.Password)
	if err != nil {
		return "", "", err
	}

	id, err := uc.UserRepo.Create(ctx, user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := uc.TokenService.GenerateRefreshToken(id)
	if err != nil {
		return "", "", err
	}

	err = uc.TokenService.StoreRefreshToken(ctx, id, refreshToken)
	if err != nil {
		return "", "", err
	}

	accessToken, err := uc.TokenService.GenerateAccessToken(id)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
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
		return "", "", err
	}

	refreshToken, err := uc.TokenService.GenerateRefreshToken(existingUser.ID)
	if err != nil {
		return "", "", err
	}

	err = uc.TokenService.StoreRefreshToken(ctx, existingUser.ID, refreshToken)
	if err != nil {
		return "", "", err
	}

	accessToken, err := uc.TokenService.GenerateAccessToken(existingUser.ID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (uc *AuthUseCase) LogoutUser(ctx context.Context, userID int64) error {

	err := uc.TokenService.RemoveRefreshToken(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (uc *AuthUseCase) GetAuthUserProfile(ctx context.Context, authUserId int64) (*entity.User, error) {

	user, err := uc.UserRepo.GetById(ctx, authUserId)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *AuthUseCase) UpdateAuthUserProfile(ctx context.Context, authUserId int64, user *entity.User) error {

	err := uc.UserRepo.UpdateById(ctx, authUserId, user)
	if err != nil {
		return err
	}

	return nil
}

func (uc *AuthUseCase) RefreshSession(ctx context.Context, refreshToken string) (string, string, error) {

	userID, err := uc.TokenService.VerifyRefreshToken(ctx, refreshToken)
	if err != nil {
		slog.Error("Check user error", slog.Any("err", err))
		return "", "", err
	}

	parsedUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		slog.Error("Check user error", slog.Any("err", err))
		return "", "", err
	}

	storedToken, err := uc.TokenService.GetRefreshToken(ctx, parsedUserID)
	if err != nil {
		slog.Error("Check user error", slog.Any("err", err))

		return "", "", err
	}

	if storedToken != refreshToken {
		slog.Error("Check user error", slog.Any("err", err))

		return "", "", err
	}

	newRefreshToken, err := uc.TokenService.GenerateRefreshToken(parsedUserID)
	if err != nil {
		slog.Error("Check user error", slog.Any("err", err))
		return "", "", err
	}

	newAccessToken, err := uc.TokenService.GenerateAccessToken(parsedUserID)
	if err != nil {
		slog.Error("Check user error", slog.Any("err", err))
		return "", "", err
	}

	err = uc.TokenService.StoreRefreshToken(ctx, parsedUserID, newRefreshToken)
	if err != nil {
		slog.Error("Check user error", slog.Any("err", err))
		return "", "", err
	}

	return newRefreshToken, newAccessToken, nil
}

func (uc *AuthUseCase) GetPublicKey() ([]byte, error) {

	pubKey, err := uc.TokenService.GetPublicKeyPEM()
	if err != nil {
		return []byte{}, err
	}

	return pubKey, nil
}
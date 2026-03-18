package usecase

import (
	"context"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
	"github.com/Prashant2307200/auth-service/internal/utils"
)

type UserUseCase struct {
	Repo interfaces.UserRepo
}

func NewUserUseCase(r interfaces.UserRepo) *UserUseCase {
	return &UserUseCase{Repo: r}
}

func (uc *UserUseCase) requireAdmin(ctx context.Context, currentUserID int64) error {
	user, err := uc.Repo.GetById(ctx, currentUserID)
	if err != nil {
		return err
	}
	if user.Role != entity.RoleAdmin {
		return utils.ErrForbidden
	}
	return nil
}

func (uc *UserUseCase) requireSelfOrAdmin(ctx context.Context, currentUserID, targetID int64) error {
	if currentUserID == targetID {
		return nil
	}
	return uc.requireAdmin(ctx, currentUserID)
}

func (uc *UserUseCase) GetUsers(ctx context.Context, currentUserID int64) ([]*entity.User, error) {
	if err := uc.requireAdmin(ctx, currentUserID); err != nil {
		return nil, err
	}
	return uc.Repo.List(ctx)
}

func (uc *UserUseCase) UpdateUserById(ctx context.Context, currentUserID, id int64, user *entity.User) error {
	if err := uc.requireSelfOrAdmin(ctx, currentUserID, id); err != nil {
		return err
	}
	return uc.Repo.UpdateById(ctx, id, user)
}

func (uc *UserUseCase) GetUserById(ctx context.Context, currentUserID, id int64) (*entity.User, error) {
	if err := uc.requireSelfOrAdmin(ctx, currentUserID, id); err != nil {
		return nil, err
	}
	return uc.Repo.GetById(ctx, id)
}

func (uc *UserUseCase) SearchUsers(ctx context.Context, currentUserId int64, search string) ([]*entity.User, error) {
	return uc.Repo.Search(ctx, currentUserId, search)
}

func (uc *UserUseCase) DeleteUserById(ctx context.Context, currentUserID, id int64) error {
	if err := uc.requireSelfOrAdmin(ctx, currentUserID, id); err != nil {
		return err
	}
	return uc.Repo.DeleteById(ctx, id)
}

func (uc *UserUseCase) CreateUser(ctx context.Context, currentUserID int64, user *entity.User) error {
	if err := uc.requireAdmin(ctx, currentUserID); err != nil {
		return err
	}
	_, err := uc.Repo.Create(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

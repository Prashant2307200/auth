package usecase

import (
	"context"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
)


type UserUseCase struct {
	Repo interfaces.UserRepo
}

func NewUserUseCase(r interfaces.UserRepo) *UserUseCase {
	return &UserUseCase{Repo: r}
}

func (uc *UserUseCase) GetUsers(ctx context.Context) ([]*entity.User, error) {
	return uc.Repo.List(ctx)
}

func (uc *UserUseCase) UpdateUserById(ctx context.Context, id int64, user *entity.User) error {
	return uc.Repo.UpdateById(ctx, id, user)
}

func (uc *UserUseCase) GetUserById(ctx context.Context, id int64) (*entity.User, error) {
	return uc.Repo.GetById(ctx, id)
}

func (uc *UserUseCase) SearchUsers(ctx context.Context, currentUserId int64, search string) ([]*entity.User, error) {
	return uc.Repo.Search(ctx, currentUserId, search)
}

func (uc *UserUseCase) DeleteUserById(ctx context.Context, id int64) error {
	return uc.Repo.DeleteById(ctx, id)
}

func (uc *UserUseCase) CreateUser(ctx context.Context, user *entity.User) error {

	_, err := uc.Repo.Create(ctx, user)
	if err != nil {
		return err
	}

	return nil
}
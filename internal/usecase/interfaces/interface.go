package interfaces

import (
	"context"
	"mime/multipart"

	"github.com/Prashant2307200/auth-service/internal/entity"
)

type UserRepo interface {
	List(ctx context.Context) ([]*entity.User, error)
	GetById(ctx context.Context, id int64) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateById(ctx context.Context, id int64, user *entity.User) error
	DeleteById(ctx context.Context, id int64) error
	Create(ctx context.Context, user *entity.User) (int64, error)
	Search(ctx context.Context, currentID int64, search string) ([]*entity.User, error)
}

type TokenService interface {
	GenerateAccessToken(userID int64) (string, error)
	GenerateRefreshToken(userID int64) (string, error)
	StoreRefreshToken(ctx context.Context, userID int64, token string) error
	RemoveRefreshToken(ctx context.Context, userID int64) error
	VerifyRefreshToken(ctx context.Context, tokenStr string) (string, error) 
	GetRefreshToken(ctx context.Context, userID int64) (string, error) 
	GetPublicKeyPEM() ([]byte, error)
}

type CloudService interface {
	UploadImage(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, error)
}
package interfaces

import (
	"context"

	"github.com/Prashant2307200/auth-service/internal/entity"
)

type UserRepo interface {
	List(ctx context.Context) ([]*entity.User, error)
	GetById(ctx context.Context, id int64) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	UpdateById(ctx context.Context, id int64, user *entity.User) error
	// UpdatePassword updates only the stored password hash for a user
	UpdatePassword(ctx context.Context, id int64, hashedPassword string) error
	DeleteById(ctx context.Context, id int64) error
	Create(ctx context.Context, user *entity.User) (int64, error)
	Search(ctx context.Context, currentID int64, search string) ([]*entity.User, error)
}

type BusinessRepo interface {
	Create(ctx context.Context, business *entity.Business) (int64, error)
	CreateWithOwner(ctx context.Context, business *entity.Business, ownerID int64) (int64, error)
	GetById(ctx context.Context, id int64) (*entity.Business, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Business, error)
	GetByOwnerId(ctx context.Context, ownerId int64) ([]*entity.Business, error)
	Update(ctx context.Context, id int64, business *entity.Business) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]*entity.Business, error)
	AddUser(ctx context.Context, businessID int64, userID int64, role int) error
	AddUserIfNotExists(ctx context.Context, businessID int64, userID int64, role int) error
	RemoveUser(ctx context.Context, businessID int64, userID int64) error
	GetUsers(ctx context.Context, businessID int64) ([]*entity.User, error)
	GetUserBusinesses(ctx context.Context, userID int64) ([]*entity.Business, error)
	GetUserRole(ctx context.Context, businessID int64, userID int64) (int, error)
	HasMembership(ctx context.Context, businessID int64, userID int64) (bool, error)

	CreateInvite(ctx context.Context, invite *entity.BusinessInvite) (int64, error)
	GetInviteByToken(ctx context.Context, token string) (*entity.BusinessInvite, error)
	RevokeInvite(ctx context.Context, inviteID int64, businessID int64) error
	AcceptInvite(ctx context.Context, inviteID int64) error
	ListInvites(ctx context.Context, businessID int64) ([]*entity.BusinessInvite, error)

	CreateDomain(ctx context.Context, domain *entity.BusinessDomain) (int64, error)
	GetDomain(ctx context.Context, businessID int64, domain string) (*entity.BusinessDomain, error)
	GetDomainByVerificationToken(ctx context.Context, token string) (*entity.BusinessDomain, error)
	FindAutoJoinBusinessByEmailDomain(ctx context.Context, emailDomain string) (*entity.Business, error)
	VerifyDomain(ctx context.Context, domainID int64) error
	UpdateDomainAutoJoin(ctx context.Context, domainID int64, businessID int64, enabled bool) error
}

type TokenService interface {
	GenerateAccessToken(userID int64, businessID ...int64) (string, error)
	GenerateRefreshToken(userID int64) (string, error)
	StoreRefreshToken(ctx context.Context, userID int64, token string) error
	RemoveRefreshToken(ctx context.Context, userID int64) error
	VerifyRefreshToken(ctx context.Context, tokenStr string) (string, error)
	VerifyToken(ctx context.Context, tokenStr string) (int64, error)
	GetRefreshToken(ctx context.Context, userID int64) (string, error)
	GetPublicKeyPEM() ([]byte, error)
}

type CloudService interface {
	GenerateUploadSignature(ctx context.Context, userID int64) (*UploadSignature, error)
}

type UploadSignature struct {
	UploadURL string `json:"upload_url"`
	APIKey    string `json:"api_key"`
	Signature string `json:"signature"`
	Timestamp string `json:"timestamp"`
	Folder    string `json:"folder"`
	PublicID  string `json:"public_id"`
	CloudName string `json:"cloud_name"`
}

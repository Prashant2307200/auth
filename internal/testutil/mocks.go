package testutil

import (
	"context"
	"errors"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
	"github.com/stretchr/testify/mock"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrCannotRevoke = errors.New("cannot revoke invitation")
)

// MockUserRepo is a mock implementation of UserRepo interface
type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) List(ctx context.Context) ([]*entity.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockUserRepo) GetById(ctx context.Context, id int64) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepo) UpdateById(ctx context.Context, id int64, user *entity.User) error {
	args := m.Called(ctx, id, user)
	return args.Error(0)
}

func (m *MockUserRepo) UpdatePassword(ctx context.Context, id int64, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepo) DeleteById(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepo) Create(ctx context.Context, user *entity.User) (int64, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepo) Search(ctx context.Context, currentID int64, search string) ([]*entity.User, error) {
	args := m.Called(ctx, currentID, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

// MockTokenService is a mock implementation of TokenService interface
type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateAccessToken(userID int64, businessID ...int64) (string, error) {
	callArgs := []interface{}{userID}
	for _, id := range businessID {
		callArgs = append(callArgs, id)
	}
	args := m.Called(callArgs...)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) GenerateRefreshToken(userID int64) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) StoreRefreshToken(ctx context.Context, userID int64, token string) error {
	args := m.Called(ctx, userID, token)
	return args.Error(0)
}

func (m *MockTokenService) RemoveRefreshToken(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTokenService) VerifyRefreshToken(ctx context.Context, tokenStr string) (string, error) {
	args := m.Called(ctx, tokenStr)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) VerifyToken(ctx context.Context, tokenStr string) (int64, error) {
	args := m.Called(ctx, tokenStr)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTokenService) GetRefreshToken(ctx context.Context, userID int64) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) GetPublicKeyPEM() ([]byte, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// MockCloudService is a mock implementation of CloudService interface
type MockCloudService struct {
	mock.Mock
}

func (m *MockCloudService) GenerateUploadSignature(ctx context.Context, userID int64) (*interfaces.UploadSignature, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*interfaces.UploadSignature), args.Error(1)
}

// MockBusinessRepo is a mock implementation of BusinessRepo interface
type MockBusinessRepo struct {
	mock.Mock
}

// MockMemberRepo is a mock for member repository
type MockMemberRepo struct{ mock.Mock }

func (m *MockMemberRepo) Create(ctx context.Context, member *entity.BusinessMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}
func (m *MockMemberRepo) GetByID(ctx context.Context, id int64) (*entity.BusinessMember, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.BusinessMember), args.Error(1)
}
func (m *MockMemberRepo) GetByUserAndBusiness(ctx context.Context, userID, businessID int64) (*entity.BusinessMember, error) {
	args := m.Called(ctx, userID, businessID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.BusinessMember), args.Error(1)
}
func (m *MockMemberRepo) ListByBusiness(ctx context.Context, businessID int64) ([]*entity.BusinessMember, error) {
	args := m.Called(ctx, businessID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.BusinessMember), args.Error(1)
}
func (m *MockMemberRepo) ListByUser(ctx context.Context, userID int64) ([]*entity.BusinessMember, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.BusinessMember), args.Error(1)
}
func (m *MockMemberRepo) Update(ctx context.Context, member *entity.BusinessMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}
func (m *MockMemberRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockMemberRepo) GetByInviteToken(ctx context.Context, token string) (*entity.BusinessMember, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.BusinessMember), args.Error(1)
}

// MockAuditRepo is a mock for AuditRepository
type MockAuditRepo struct{ mock.Mock }

func (m *MockAuditRepo) Log(ctx context.Context, audit *entity.AuditLog) error {
	args := m.Called(ctx, audit)
	return args.Error(0)
}
func (m *MockAuditRepo) GetByID(ctx context.Context, id int64) (*entity.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AuditLog), args.Error(1)
}
func (m *MockAuditRepo) ListByBusiness(ctx context.Context, businessID int64, limit, offset int) ([]*entity.AuditLog, error) {
	args := m.Called(ctx, businessID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditLog), args.Error(1)
}
func (m *MockAuditRepo) ListByUser(ctx context.Context, businessID, userID int64, limit, offset int) ([]*entity.AuditLog, error) {
	args := m.Called(ctx, businessID, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditLog), args.Error(1)
}
func (m *MockAuditRepo) Export(ctx context.Context, businessID int64) ([]*entity.AuditLog, error) {
	args := m.Called(ctx, businessID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditLog), args.Error(1)
}

func (m *MockBusinessRepo) Create(ctx context.Context, business *entity.Business) (int64, error) {
	args := m.Called(ctx, business)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBusinessRepo) CreateWithOwner(ctx context.Context, business *entity.Business, ownerID int64) (int64, error) {
	args := m.Called(ctx, business, ownerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBusinessRepo) GetById(ctx context.Context, id int64) (*entity.Business, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Business), args.Error(1)
}

func (m *MockBusinessRepo) GetBySlug(ctx context.Context, slug string) (*entity.Business, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Business), args.Error(1)
}

func (m *MockBusinessRepo) GetByOwnerId(ctx context.Context, ownerId int64) ([]*entity.Business, error) {
	args := m.Called(ctx, ownerId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Business), args.Error(1)
}

func (m *MockBusinessRepo) Update(ctx context.Context, id int64, business *entity.Business) error {
	args := m.Called(ctx, id, business)
	return args.Error(0)
}

func (m *MockBusinessRepo) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBusinessRepo) List(ctx context.Context) ([]*entity.Business, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Business), args.Error(1)
}

func (m *MockBusinessRepo) AddUser(ctx context.Context, businessID int64, userID int64, role int) error {
	args := m.Called(ctx, businessID, userID, role)
	return args.Error(0)
}

func (m *MockBusinessRepo) RemoveUser(ctx context.Context, businessID int64, userID int64) error {
	args := m.Called(ctx, businessID, userID)
	return args.Error(0)
}

func (m *MockBusinessRepo) GetUsers(ctx context.Context, businessID int64) ([]*entity.User, error) {
	args := m.Called(ctx, businessID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockBusinessRepo) GetUserBusinesses(ctx context.Context, userID int64) ([]*entity.Business, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Business), args.Error(1)
}

func (m *MockBusinessRepo) GetUserRole(ctx context.Context, businessID int64, userID int64) (int, error) {
	args := m.Called(ctx, businessID, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockBusinessRepo) AddUserIfNotExists(ctx context.Context, businessID int64, userID int64, role int) error {
	args := m.Called(ctx, businessID, userID, role)
	return args.Error(0)
}

func (m *MockBusinessRepo) HasMembership(ctx context.Context, businessID int64, userID int64) (bool, error) {
	args := m.Called(ctx, businessID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockBusinessRepo) CreateInvite(ctx context.Context, invite *entity.BusinessInvite) (int64, error) {
	args := m.Called(ctx, invite)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBusinessRepo) GetInviteByToken(ctx context.Context, token string) (*entity.BusinessInvite, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.BusinessInvite), args.Error(1)
}

func (m *MockBusinessRepo) RevokeInvite(ctx context.Context, inviteID int64, businessID int64) error {
	args := m.Called(ctx, inviteID, businessID)
	return args.Error(0)
}

func (m *MockBusinessRepo) AcceptInvite(ctx context.Context, inviteID int64) error {
	args := m.Called(ctx, inviteID)
	return args.Error(0)
}

func (m *MockBusinessRepo) ListInvites(ctx context.Context, businessID int64) ([]*entity.BusinessInvite, error) {
	args := m.Called(ctx, businessID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.BusinessInvite), args.Error(1)
}

func (m *MockBusinessRepo) CreateDomain(ctx context.Context, domain *entity.BusinessDomain) (int64, error) {
	args := m.Called(ctx, domain)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBusinessRepo) GetDomain(ctx context.Context, businessID int64, domain string) (*entity.BusinessDomain, error) {
	args := m.Called(ctx, businessID, domain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.BusinessDomain), args.Error(1)
}

func (m *MockBusinessRepo) GetDomainByVerificationToken(ctx context.Context, token string) (*entity.BusinessDomain, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.BusinessDomain), args.Error(1)
}

func (m *MockBusinessRepo) FindAutoJoinBusinessByEmailDomain(ctx context.Context, emailDomain string) (*entity.Business, error) {
	args := m.Called(ctx, emailDomain)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Business), args.Error(1)
}

func (m *MockBusinessRepo) VerifyDomain(ctx context.Context, domainID int64) error {
	args := m.Called(ctx, domainID)
	return args.Error(0)
}

func (m *MockBusinessRepo) UpdateDomainAutoJoin(ctx context.Context, domainID int64, businessID int64, enabled bool) error {
	args := m.Called(ctx, domainID, businessID, enabled)
	return args.Error(0)
}

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendInvite(ctx context.Context, to string, token string) error {
	args := m.Called(ctx, to, token)
	return args.Error(0)
}

type MockTeamUsecase struct {
	mock.Mock
}

func (m *MockTeamUsecase) InviteUser(ctx context.Context, businessID int64, email string, role int) (string, error) {
	args := m.Called(ctx, businessID, email, role)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockTeamUsecase) AcceptInvitation(ctx context.Context, inviteToken string) error {
	args := m.Called(ctx, inviteToken)
	return args.Error(0)
}

func (m *MockTeamUsecase) RevokeInvitation(ctx context.Context, inviteToken string) error {
	args := m.Called(ctx, inviteToken)
	return args.Error(0)
}

func (m *MockTeamUsecase) ListMembers(ctx context.Context, businessID int64) ([]*entity.BusinessMember, error) {
	args := m.Called(ctx, businessID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.BusinessMember), args.Error(1)
}

func (m *MockTeamUsecase) RemoveMember(ctx context.Context, businessID int64, memberID int64) error {
	args := m.Called(ctx, businessID, memberID)
	return args.Error(0)
}

func (m *MockTeamUsecase) UpdateMemberRole(ctx context.Context, businessID int64, memberID int64, newRole int) error {
	args := m.Called(ctx, businessID, memberID, newRole)
	return args.Error(0)
}

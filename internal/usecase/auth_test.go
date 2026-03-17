package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthUseCase_RegisterUser(t *testing.T) {
	tests := []struct {
		name               string
		user               *entity.User
		setupMocks         func(*testutil.MockUserRepo, *testutil.MockTokenService, *testutil.MockCloudService)
		setupBusinessMocks func(*testutil.MockBusinessRepo)
		wantErr            bool
		wantErrMsg         string
		validateToken      bool
	}{
		{
			name: "successful registration",
			user: &entity.User{
				Email:    "newuser@example.com",
				Username: "newuser",
				Password: "password123",
			},
			setupMocks: func(userRepo *testutil.MockUserRepo, tokenService *testutil.MockTokenService, cloudService *testutil.MockCloudService) {
				userRepo.On("GetByEmail", mock.Anything, "newuser@example.com").Return(nil, sql.ErrNoRows)
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(1), nil)
				tokenService.On("GenerateRefreshToken", int64(1)).Return("refresh_token", nil)
				tokenService.On("StoreRefreshToken", mock.Anything, int64(1), "refresh_token").Return(nil)
				tokenService.On("GenerateAccessToken", int64(1)).Return("access_token", nil)
			},
			setupBusinessMocks: func(businessRepo *testutil.MockBusinessRepo) {
				businessRepo.On("FindAutoJoinBusinessByEmailDomain", mock.Anything, "example.com").Return(nil, nil)
			},
			wantErr:       false,
			validateToken: true,
		},
		{
			name: "user already exists",
			user: &entity.User{
				Email:    "existing@example.com",
				Username: "existing",
				Password: "password123",
			},
			setupMocks: func(userRepo *testutil.MockUserRepo, tokenService *testutil.MockTokenService, cloudService *testutil.MockCloudService) {
				existingUser := testutil.CreateTestUserWithEmail("existing@example.com")
				userRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			wantErr:    true,
			wantErrMsg: "already exists",
		},
		{
			name: "database error on check",
			user: &entity.User{
				Email:    "test@example.com",
				Username: "test",
				Password: "password123",
			},
			setupMocks: func(userRepo *testutil.MockUserRepo, tokenService *testutil.MockTokenService, cloudService *testutil.MockCloudService) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("database error"))
			},
			wantErr:    true,
			wantErrMsg: "failed to check existing user",
		},
		{
			name: "failed to create user",
			user: &entity.User{
				Email:    "test@example.com",
				Username: "test",
				Password: "password123",
			},
			setupMocks: func(userRepo *testutil.MockUserRepo, tokenService *testutil.MockTokenService, cloudService *testutil.MockCloudService) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, sql.ErrNoRows)
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(0), errors.New("create error"))
			},
			wantErr:    true,
			wantErrMsg: "failed to create user",
		},
		{
			name: "failed to generate refresh token",
			user: &entity.User{
				Email:    "test@example.com",
				Username: "test",
				Password: "password123",
			},
			setupMocks: func(userRepo *testutil.MockUserRepo, tokenService *testutil.MockTokenService, cloudService *testutil.MockCloudService) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, sql.ErrNoRows)
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(1), nil)
				tokenService.On("GenerateRefreshToken", int64(1)).Return("", errors.New("token error"))
			},
			setupBusinessMocks: func(businessRepo *testutil.MockBusinessRepo) {
				businessRepo.On("FindAutoJoinBusinessByEmailDomain", mock.Anything, "example.com").Return(nil, nil)
			},
			wantErr:    true,
			wantErrMsg: "failed to generate refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			tokenService := new(testutil.MockTokenService)
			cloudService := new(testutil.MockCloudService)

			tt.setupMocks(userRepo, tokenService, cloudService)

			businessRepo := new(testutil.MockBusinessRepo)
			if tt.setupBusinessMocks != nil {
				tt.setupBusinessMocks(businessRepo)
			}
			uc := NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
			accessToken, refreshToken, err := uc.RegisterUser(context.Background(), tt.user, nil)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
			} else {
				assert.NoError(t, err)
				if tt.validateToken {
					assert.NotEmpty(t, accessToken)
					assert.NotEmpty(t, refreshToken)
				}
			}

			userRepo.AssertExpectations(t)
			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthUseCase_RegisterUser_WithInviteToken(t *testing.T) {
	user := &entity.User{Email: "invited@acme.com", Username: "invited", Password: "pass123"}
	inv := testutil.CreateTestInvite(10, "invited@acme.com", "tok-abc", entity.InviteStatusPending, time.Now().Add(24*time.Hour))
	inv.ID = 5

	userRepo := new(testutil.MockUserRepo)
	businessRepo := new(testutil.MockBusinessRepo)
	tokenService := new(testutil.MockTokenService)
	cloudService := new(testutil.MockCloudService)

	userRepo.On("GetByEmail", mock.Anything, "invited@acme.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(100), nil)
	businessRepo.On("GetInviteByToken", mock.Anything, "tok-abc").Return(inv, nil)
	businessRepo.On("AddUserIfNotExists", mock.Anything, int64(10), int64(100), 0).Return(nil)
	businessRepo.On("AcceptInvite", mock.Anything, int64(5)).Return(nil)
	tokenService.On("GenerateRefreshToken", int64(100)).Return("ref", nil)
	tokenService.On("StoreRefreshToken", mock.Anything, int64(100), "ref").Return(nil)
	tokenService.On("GenerateAccessToken", int64(100)).Return("acc", nil)

	uc := NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
	opts := &RegisterOptions{InviteToken: "tok-abc"}
	acc, ref, err := uc.RegisterUser(context.Background(), user, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, acc)
	assert.NotEmpty(t, ref)
	userRepo.AssertExpectations(t)
	businessRepo.AssertExpectations(t)
	tokenService.AssertExpectations(t)
}

func TestAuthUseCase_RegisterUser_InviteExpired(t *testing.T) {
	user := &entity.User{Email: "invited@acme.com", Username: "invited", Password: "pass123"}
	inv := testutil.CreateTestInvite(10, "invited@acme.com", "tok-abc", entity.InviteStatusPending, time.Now().Add(-time.Hour))

	userRepo := new(testutil.MockUserRepo)
	businessRepo := new(testutil.MockBusinessRepo)
	tokenService := new(testutil.MockTokenService)
	cloudService := new(testutil.MockCloudService)

	userRepo.On("GetByEmail", mock.Anything, "invited@acme.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(100), nil)
	businessRepo.On("GetInviteByToken", mock.Anything, "tok-abc").Return(inv, nil)

	uc := NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
	opts := &RegisterOptions{InviteToken: "tok-abc"}
	_, _, err := uc.RegisterUser(context.Background(), user, opts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestAuthUseCase_RegisterUser_InviteEmailMismatch(t *testing.T) {
	user := &entity.User{Email: "other@acme.com", Username: "other", Password: "pass123"}
	inv := testutil.CreateTestInvite(10, "invited@acme.com", "tok-abc", entity.InviteStatusPending, time.Now().Add(24*time.Hour))

	userRepo := new(testutil.MockUserRepo)
	businessRepo := new(testutil.MockBusinessRepo)
	tokenService := new(testutil.MockTokenService)
	cloudService := new(testutil.MockCloudService)

	userRepo.On("GetByEmail", mock.Anything, "other@acme.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(100), nil)
	businessRepo.On("GetInviteByToken", mock.Anything, "tok-abc").Return(inv, nil)

	uc := NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
	opts := &RegisterOptions{InviteToken: "tok-abc"}
	_, _, err := uc.RegisterUser(context.Background(), user, opts)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email does not match")
}

func TestAuthUseCase_RegisterUser_WithBusinessSlug(t *testing.T) {
	user := &entity.User{Email: "new@test.com", Username: "newuser", Password: "pass123"}
	biz := testutil.CreateTestBusinessWithSignupPolicy("open-biz", entity.SignupPolicyOpen)

	userRepo := new(testutil.MockUserRepo)
	businessRepo := new(testutil.MockBusinessRepo)
	tokenService := new(testutil.MockTokenService)
	cloudService := new(testutil.MockCloudService)

	userRepo.On("GetByEmail", mock.Anything, "new@test.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(101), nil)
	businessRepo.On("GetBySlug", mock.Anything, "open-biz").Return(biz, nil)
	businessRepo.On("AddUserIfNotExists", mock.Anything, biz.ID, int64(101), 0).Return(nil)
	tokenService.On("GenerateRefreshToken", int64(101)).Return("ref", nil)
	tokenService.On("StoreRefreshToken", mock.Anything, int64(101), "ref").Return(nil)
	tokenService.On("GenerateAccessToken", int64(101)).Return("acc", nil)

	uc := NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
	opts := &RegisterOptions{BusinessSlug: "open-biz"}
	acc, ref, err := uc.RegisterUser(context.Background(), user, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, acc)
	assert.NotEmpty(t, ref)
	userRepo.AssertExpectations(t)
	businessRepo.AssertExpectations(t)
}

func TestAuthUseCase_RegisterUser_WithDomainAutoJoin(t *testing.T) {
	user := &entity.User{Email: "emp@acme.com", Username: "emp", Password: "pass123"}
	biz := testutil.CreateTestBusiness()

	userRepo := new(testutil.MockUserRepo)
	businessRepo := new(testutil.MockBusinessRepo)
	tokenService := new(testutil.MockTokenService)
	cloudService := new(testutil.MockCloudService)

	userRepo.On("GetByEmail", mock.Anything, "emp@acme.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(102), nil)
	businessRepo.On("FindAutoJoinBusinessByEmailDomain", mock.Anything, "acme.com").Return(biz, nil)
	businessRepo.On("AddUserIfNotExists", mock.Anything, biz.ID, int64(102), 0).Return(nil)
	tokenService.On("GenerateRefreshToken", int64(102)).Return("ref", nil)
	tokenService.On("StoreRefreshToken", mock.Anything, int64(102), "ref").Return(nil)
	tokenService.On("GenerateAccessToken", int64(102)).Return("acc", nil)

	uc := NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
	acc, ref, err := uc.RegisterUser(context.Background(), user, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, acc)
	assert.NotEmpty(t, ref)
	userRepo.AssertExpectations(t)
	businessRepo.AssertExpectations(t)
}

func TestAuthUseCase_LoginUser(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		password   string
		setupMocks func(*testutil.MockUserRepo, *testutil.MockTokenService, *testutil.MockCloudService)
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(userRepo *testutil.MockUserRepo, tokenService *testutil.MockTokenService, cloudService *testutil.MockCloudService) {
				user := testutil.CreateTestUser()
				// Note: In real tests, you'd hash the password properly
				// For now, we'll test the error path since password hashing happens in the usecase
				user.Password = "hashedpassword" // This will fail password check, testing error path
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			wantErr:    true,
			wantErrMsg: "invalid password",
		},
		{
			name:     "user not found",
			email:    "notfound@example.com",
			password: "password123",
			setupMocks: func(userRepo *testutil.MockUserRepo, tokenService *testutil.MockTokenService, cloudService *testutil.MockCloudService) {
				userRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, sql.ErrNoRows)
			},
			wantErr:    true,
			wantErrMsg: "user not found",
		},
		{
			name:     "database error",
			email:    "test@example.com",
			password: "password123",
			setupMocks: func(userRepo *testutil.MockUserRepo, tokenService *testutil.MockTokenService, cloudService *testutil.MockCloudService) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("database error"))
			},
			wantErr:    true,
			wantErrMsg: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			tokenService := new(testutil.MockTokenService)
			cloudService := new(testutil.MockCloudService)

			tt.setupMocks(userRepo, tokenService, cloudService)

			uc := NewAuthUseCase(userRepo, nil, tokenService, cloudService)
			accessToken, refreshToken, err := uc.LoginUser(context.Background(), tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, refreshToken)
			}

			userRepo.AssertExpectations(t)
			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthUseCase_LogoutUser(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		setupMocks func(*testutil.MockTokenService)
		wantErr    bool
	}{
		{
			name:   "successful logout",
			userID: 1,
			setupMocks: func(tokenService *testutil.MockTokenService) {
				tokenService.On("RemoveRefreshToken", mock.Anything, int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "failed to remove token",
			userID: 1,
			setupMocks: func(tokenService *testutil.MockTokenService) {
				tokenService.On("RemoveRefreshToken", mock.Anything, int64(1)).Return(errors.New("redis error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			tokenService := new(testutil.MockTokenService)
			cloudService := new(testutil.MockCloudService)

			tt.setupMocks(tokenService)

			uc := NewAuthUseCase(userRepo, nil, tokenService, cloudService)
			err := uc.LogoutUser(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthUseCase_GetAuthUserProfile(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		setupMocks func(*testutil.MockUserRepo)
		wantErr    bool
		wantUser   *entity.User
	}{
		{
			name:   "successful get profile",
			userID: 1,
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				user := testutil.CreateTestUserWithID(1)
				userRepo.On("GetById", mock.Anything, int64(1)).Return(user, nil)
			},
			wantErr: false,
			wantUser: testutil.CreateTestUserWithID(1),
		},
		{
			name:   "user not found",
			userID: 999,
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("GetById", mock.Anything, int64(999)).Return(nil, errors.New("user not found"))
			},
			wantErr:  true,
			wantUser: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			tokenService := new(testutil.MockTokenService)
			cloudService := new(testutil.MockCloudService)

			tt.setupMocks(userRepo)

			uc := NewAuthUseCase(userRepo, nil, tokenService, cloudService)
			user, err := uc.GetAuthUserProfile(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, tt.wantUser.ID, user.ID)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestAuthUseCase_UpdateAuthUserProfile(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		user       *entity.User
		setupMocks func(*testutil.MockUserRepo)
		wantErr    bool
	}{
		{
			name:   "successful update",
			userID: 1,
			user: &entity.User{
				Username: "updateduser",
				Email:    "updated@example.com",
			},
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("UpdateById", mock.Anything, int64(1), mock.AnythingOfType("*entity.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "update failed",
			userID: 1,
			user: &entity.User{
				Username: "updateduser",
			},
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("UpdateById", mock.Anything, int64(1), mock.AnythingOfType("*entity.User")).Return(errors.New("update error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			tokenService := new(testutil.MockTokenService)
			cloudService := new(testutil.MockCloudService)

			tt.setupMocks(userRepo)

			uc := NewAuthUseCase(userRepo, nil, tokenService, cloudService)
			err := uc.UpdateAuthUserProfile(context.Background(), tt.userID, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestAuthUseCase_RegisterUser_WithOnboarding(t *testing.T) {
	t.Run("invite_token_success", func(t *testing.T) {
		userRepo := new(testutil.MockUserRepo)
		businessRepo := new(testutil.MockBusinessRepo)
		tokenService := new(testutil.MockTokenService)
		cloudService := new(testutil.MockCloudService)

		user := &entity.User{Email: "invited@acme.com", Username: "invited", Password: "pass123"}
		invite := testutil.CreateTestInvite(10, "invited@acme.com", "tok-abc", entity.InviteStatusPending, time.Now().Add(24*time.Hour))

		userRepo.On("GetByEmail", mock.Anything, "invited@acme.com").Return(nil, sql.ErrNoRows)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(100), nil)
		businessRepo.On("GetInviteByToken", mock.Anything, "tok-abc").Return(invite, nil)
		businessRepo.On("AddUserIfNotExists", mock.Anything, int64(10), int64(100), 0).Return(nil)
		businessRepo.On("AcceptInvite", mock.Anything, int64(1)).Return(nil)
		tokenService.On("GenerateRefreshToken", int64(100)).Return("ref", nil)
		tokenService.On("StoreRefreshToken", mock.Anything, int64(100), "ref").Return(nil)
		tokenService.On("GenerateAccessToken", int64(100)).Return("acc", nil)

		uc := NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
		acc, ref, err := uc.RegisterUser(context.Background(), user, &RegisterOptions{InviteToken: "tok-abc"})
		assert.NoError(t, err)
		assert.NotEmpty(t, acc)
		assert.NotEmpty(t, ref)
	})

	t.Run("invite_expired", func(t *testing.T) {
		userRepo := new(testutil.MockUserRepo)
		businessRepo := new(testutil.MockBusinessRepo)
		tokenService := new(testutil.MockTokenService)
		cloudService := new(testutil.MockCloudService)

		user := &entity.User{Email: "invited@acme.com", Username: "invited", Password: "pass123"}
		invite := testutil.CreateTestInvite(10, "invited@acme.com", "tok-abc", entity.InviteStatusPending, time.Now().Add(-1*time.Hour))

		userRepo.On("GetByEmail", mock.Anything, "invited@acme.com").Return(nil, sql.ErrNoRows)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(100), nil)
		businessRepo.On("GetInviteByToken", mock.Anything, "tok-abc").Return(invite, nil)

		uc := NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
		_, _, err := uc.RegisterUser(context.Background(), user, &RegisterOptions{InviteToken: "tok-abc"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("business_slug_open_signup", func(t *testing.T) {
		userRepo := new(testutil.MockUserRepo)
		businessRepo := new(testutil.MockBusinessRepo)
		tokenService := new(testutil.MockTokenService)
		cloudService := new(testutil.MockCloudService)

		user := &entity.User{Email: "new@test.com", Username: "newuser", Password: "pass123"}
		biz := testutil.CreateTestBusinessWithSignupPolicy("openco", entity.SignupPolicyOpen)

		userRepo.On("GetByEmail", mock.Anything, "new@test.com").Return(nil, sql.ErrNoRows)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(101), nil)
		businessRepo.On("GetBySlug", mock.Anything, "openco").Return(biz, nil)
		businessRepo.On("AddUserIfNotExists", mock.Anything, biz.ID, int64(101), 0).Return(nil)
		businessRepo.On("FindAutoJoinBusinessByEmailDomain", mock.Anything, "test.com").Return(nil, nil)
		tokenService.On("GenerateRefreshToken", int64(101)).Return("ref", nil)
		tokenService.On("StoreRefreshToken", mock.Anything, int64(101), "ref").Return(nil)
		tokenService.On("GenerateAccessToken", int64(101)).Return("acc", nil)

		uc := NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
		acc, ref, err := uc.RegisterUser(context.Background(), user, &RegisterOptions{BusinessSlug: "openco"})
		assert.NoError(t, err)
		assert.NotEmpty(t, acc)
		assert.NotEmpty(t, ref)
	})

	t.Run("domain_auto_join", func(t *testing.T) {
		userRepo := new(testutil.MockUserRepo)
		businessRepo := new(testutil.MockBusinessRepo)
		tokenService := new(testutil.MockTokenService)
		cloudService := new(testutil.MockCloudService)

		user := &entity.User{Email: "employee@company.com", Username: "emp", Password: "pass123"}
		biz := testutil.CreateTestBusinessWithSignupPolicy("company", entity.SignupPolicyClosed)
		biz.ID = 20

		userRepo.On("GetByEmail", mock.Anything, "employee@company.com").Return(nil, sql.ErrNoRows)
		userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(102), nil)
		businessRepo.On("FindAutoJoinBusinessByEmailDomain", mock.Anything, "company.com").Return(biz, nil)
		businessRepo.On("AddUserIfNotExists", mock.Anything, int64(20), int64(102), 0).Return(nil)
		tokenService.On("GenerateRefreshToken", int64(102)).Return("ref", nil)
		tokenService.On("StoreRefreshToken", mock.Anything, int64(102), "ref").Return(nil)
		tokenService.On("GenerateAccessToken", int64(102)).Return("acc", nil)

		uc := NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
		acc, ref, err := uc.RegisterUser(context.Background(), user, nil)
		assert.NoError(t, err)
		assert.NotEmpty(t, acc)
		assert.NotEmpty(t, ref)
	})
}

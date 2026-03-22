package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserUseCase_GetUsers(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*testutil.MockUserRepo)
		wantErr    bool
		wantCount  int
	}{
		{
			name: "successful get users",
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("GetById", mock.Anything, int64(99)).Return(testutil.CreateTestAdminWithID(99), nil)
				userRepo.On("List", mock.Anything).Return(testutil.CreateTestUserList(1, 2), nil)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "empty list",
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("GetById", mock.Anything, int64(99)).Return(testutil.CreateTestAdminWithID(99), nil)
				userRepo.On("List", mock.Anything).Return([]*entity.User{}, nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "database error",
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("GetById", mock.Anything, int64(99)).Return(testutil.CreateTestAdminWithID(99), nil)
				userRepo.On("List", mock.Anything).Return(nil, errors.New("database error"))
			},
			wantErr: true,
		},
		{
			name: "forbidden when not admin",
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("GetById", mock.Anything, int64(99)).Return(testutil.CreateTestUserWithID(99), nil)
			},
			wantErr: true,
		},
	}

	const adminID int64 = 99
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			tt.setupMocks(userRepo)

			uc := NewUserUseCase(userRepo)
			users, err := uc.GetUsers(context.Background(), adminID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, users)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, users)
				assert.Len(t, users, tt.wantCount)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestUserUseCase_GetUserById(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		setupMocks func(*testutil.MockUserRepo)
		wantErr    bool
		wantUserID int64
	}{
		{
			name:   "successful get user (self)",
			userID: 1,
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				user := testutil.CreateTestUserWithID(1)
				userRepo.On("GetById", mock.Anything, int64(1)).Return(user, nil)
			},
			wantErr:    false,
			wantUserID: 1,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("GetById", mock.Anything, int64(999)).Return(nil, errors.New("user not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			tt.setupMocks(userRepo)

			uc := NewUserUseCase(userRepo)
			user, err := uc.GetUserById(context.Background(), tt.userID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.wantUserID, user.ID)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestUserUseCase_CreateUser(t *testing.T) {
	tests := []struct {
		name       string
		user       *entity.User
		setupMocks func(*testutil.MockUserRepo)
		wantErr    bool
	}{
		{
			name: "successful create",
			user: testutil.CreateTestUser(),
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("GetById", mock.Anything, int64(99)).Return(testutil.CreateTestAdminWithID(99), nil)
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(1), nil)
			},
			wantErr: false,
		},
		{
			name: "create failed",
			user: testutil.CreateTestUser(),
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("GetById", mock.Anything, int64(99)).Return(testutil.CreateTestAdminWithID(99), nil)
				userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(0), errors.New("create error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			tt.setupMocks(userRepo)

			uc := NewUserUseCase(userRepo)
			err := uc.CreateUser(context.Background(), 99, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestUserUseCase_UpdateUserById(t *testing.T) {
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
			user:   testutil.CreateTestUser(),
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("UpdateById", mock.Anything, int64(1), mock.AnythingOfType("*entity.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "update failed",
			userID: 1,
			user:   testutil.CreateTestUser(),
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("UpdateById", mock.Anything, int64(1), mock.AnythingOfType("*entity.User")).Return(errors.New("update error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			tt.setupMocks(userRepo)

			uc := NewUserUseCase(userRepo)
			err := uc.UpdateUserById(context.Background(), tt.userID, tt.userID, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestUserUseCase_DeleteUserById(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		setupMocks func(*testutil.MockUserRepo)
		wantErr    bool
	}{
		{
			name:   "successful delete",
			userID: 1,
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("DeleteById", mock.Anything, int64(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "delete failed",
			userID: 1,
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("DeleteById", mock.Anything, int64(1)).Return(errors.New("delete error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			tt.setupMocks(userRepo)

			uc := NewUserUseCase(userRepo)
			err := uc.DeleteUserById(context.Background(), tt.userID, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

func TestUserUseCase_SearchUsers(t *testing.T) {
	tests := []struct {
		name          string
		currentUserID int64
		search        string
		setupMocks    func(*testutil.MockUserRepo)
		wantErr       bool
		wantCount     int
	}{
		{
			name:          "successful search",
			currentUserID: 1,
			search:        "test",
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("Search", mock.Anything, int64(1), "test").Return(testutil.CreateTestUserList(2), nil)
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:          "empty search results",
			currentUserID: 1,
			search:        "nonexistent",
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("Search", mock.Anything, int64(1), "nonexistent").Return([]*entity.User{}, nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:          "search error",
			currentUserID: 1,
			search:        "test",
			setupMocks: func(userRepo *testutil.MockUserRepo) {
				userRepo.On("Search", mock.Anything, int64(1), "test").Return(nil, errors.New("search error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := new(testutil.MockUserRepo)
			tt.setupMocks(userRepo)

			uc := NewUserUseCase(userRepo)
			users, err := uc.SearchUsers(context.Background(), tt.currentUserID, tt.search)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, users)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, users)
				assert.Len(t, users, tt.wantCount)
			}

			userRepo.AssertExpectations(t)
		})
	}
}

package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestUserRepoWithMock(t *testing.T) {
	mockRepo := &testutil.MockUserRepo{}
	uc := mockRepo

	ctx := context.Background()

	// List success
	mockRepo.On("List", ctx).Return([]*entity.User{{ID: 1, Username: "a"}}, nil)
	users, err := uc.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, users, 1)

	// GetById not found
	mockRepo.On("GetById", ctx, int64(2)).Return((*entity.User)(nil), sql.ErrNoRows)
	u, err := uc.GetById(ctx, 2)
	assert.Nil(t, u)
	assert.Error(t, err)

	// Create error when nil
	mockRepo.On("Create", ctx, (*entity.User)(nil)).Return(int64(0), errors.New("user cannot be nil"))
	id, err := uc.Create(ctx, nil)
	assert.Equal(t, int64(0), id)
	assert.Error(t, err)
}

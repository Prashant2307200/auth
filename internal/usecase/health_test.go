package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHealthCheck_OK(t *testing.T) {
	mockRepo := &testutil.MockUserRepo{}
	mockRepo.On("List", mock.Anything).Return([]*entity.User{}, nil)

	hr := NewHealthUseCase(mockRepo, nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	status, err := hr.Check(ctx)
	assert.NoError(t, err)
	assert.Contains(t, []string{"healthy", "degraded"}, status.Status)
}

package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_ServeHTTP(t *testing.T) {
	mockRepo := &testutil.MockUserRepo{}
	mockRepo.On("List", mock.Anything).Return([]*entity.User{}, nil)

	uc := usecase.NewHealthUseCase(mockRepo, nil)
	h := NewHealthHandler(uc)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	res := w.Result()
	require.Equal(t, http.StatusOK, res.StatusCode)
}

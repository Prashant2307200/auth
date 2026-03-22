package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetAllUsers_Handler_OK(t *testing.T) {
	adminID := int64(99)
	mockRepo := &testutil.MockUserRepo{}
	mockRepo.On("GetById", mock.Anything, adminID).Return(testutil.CreateTestAdminWithID(adminID), nil)
	mockRepo.On("List", mock.Anything).Return([]*entity.User{testutil.CreateTestUserWithID(1)}, nil)

	uc := usecase.NewUserUseCase(mockRepo)
	h := &UserHandler{UC: uc}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), adminID))
	rr := httptest.NewRecorder()

	h.getAll(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var body []map[string]any
	err := json.NewDecoder(rr.Body).Decode(&body)
	require.NoError(t, err)
	require.Len(t, body, 1)
}

package handler

import (
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

func TestGetAll_Handler(t *testing.T) {
	adminID := int64(99)
	m := &testutil.MockUserRepo{}
	m.On("GetById", mock.Anything, adminID).Return(testutil.CreateTestAdminWithID(adminID), nil)
	m.On("List", mock.Anything).Return([]*entity.User{testutil.CreateTestUserWithID(1)}, nil)
	uc := usecase.NewUserUseCase(m)
	h := NewUserHandler(uc)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), adminID))
	w := httptest.NewRecorder()

	h.getAll(w, req)
	require.Equal(t, http.StatusOK, w.Result().StatusCode)
}

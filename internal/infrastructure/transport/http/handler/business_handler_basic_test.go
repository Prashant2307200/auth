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

func TestBusinessGetAll_Handler_OK(t *testing.T) {
	mockBusinessRepo := &testutil.MockBusinessRepo{}
	mockUserRepo := &testutil.MockUserRepo{}
	mockBusinessRepo.On("GetUserBusinesses", mock.Anything, int64(1)).Return([]*entity.Business{testutil.CreateTestBusiness()}, nil)

	uc := usecase.NewBusinessUseCase(mockBusinessRepo, mockUserRepo)
	h := &BusinessHandler{UC: uc}
	req := httptest.NewRequest(http.MethodGet, "/businesses", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()

	h.getMyBusinesses(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	var body []map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&body)
	require.NoError(t, err)
	require.Len(t, body, 1)
}

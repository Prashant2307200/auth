package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"errors"
	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBusinessHandler_GetByID_OK(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := &BusinessHandler{UC: uc}

	mockBusiness.On("GetById", mock.Anything, int64(5)).Return(testutil.CreateTestBusinessWithID(5), nil)

	req := httptest.NewRequest(http.MethodGet, "/business?id=5", nil)
	rr := httptest.NewRecorder()
	h.getByID(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got entity.Business
	err := json.NewDecoder(rr.Body).Decode(&got)
	require.NoError(t, err)
	require.Equal(t, int64(5), got.ID)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_Update_UnauthorizedAndForbidden(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := &BusinessHandler{UC: uc}

	// no auth -> 401 (body irrelevant; use seed-based for consistency)
	body, _ := json.Marshal(testutil.CreateTestBusiness())
	req := httptest.NewRequest(http.MethodPut, "/business?id=10", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.update(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)

	// auth but insufficient role -> 403 (valid body so we reach usecase)
	validBody, _ := json.Marshal(testutil.CreateTestBusinessWithSignupPolicy("valid-co", "closed"))
	mockBusiness.On("GetUserRole", mock.Anything, int64(10), int64(2)).Return(0, nil)
	req2 := httptest.NewRequest(http.MethodPut, "/business?id=10", bytes.NewReader(validBody))
	req2 = req2.WithContext(middleware.WithUserID(req2.Context(), 2))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	h.update(rr2, req2)
	require.Equal(t, http.StatusForbidden, rr2.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_Delete_Success(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := &BusinessHandler{UC: uc}

	mockBusiness.On("GetUserRole", mock.Anything, int64(20), int64(3)).Return(usecase.BusinessRoleOwner, nil)
	mockBusiness.On("Delete", mock.Anything, int64(20)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/business?id=20", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 3))
	rr := httptest.NewRecorder()
	h.delete(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockBusiness.AssertExpectations(t)
}

func TestBusinessHandler_GetMyBusinessesAndGetUsers(t *testing.T) {
	mockBusiness := &testutil.MockBusinessRepo{}
	mockUser := &testutil.MockUserRepo{}
	uc := usecase.NewBusinessUseCase(mockBusiness, mockUser)
	h := &BusinessHandler{UC: uc}

	mockBusiness.On("GetUserBusinesses", mock.Anything, int64(4)).Return([]*entity.Business{testutil.CreateTestBusiness()}, nil)
	req := httptest.NewRequest(http.MethodGet, "/businesses", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 4))
	rr := httptest.NewRecorder()
	h.getMyBusinesses(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	// getUsers: forbidden when no role
	mockBusiness.ExpectedCalls = nil
	mockBusiness.On("GetUserRole", mock.Anything, int64(50), int64(6)).Return(0, errors.New("no"))
	req2 := httptest.NewRequest(http.MethodGet, "/business/users?id=50", nil)
	req2 = req2.WithContext(middleware.WithUserID(req2.Context(), 6))
	rr2 := httptest.NewRecorder()
	h.getUsers(rr2, req2)
	require.Equal(t, http.StatusForbidden, rr2.Code)
}

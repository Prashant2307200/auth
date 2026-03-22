package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/Prashant2307200/auth-service/pkg/db"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_ListUsers_Success(t *testing.T) {
	adminID := int64(200)
	mockUser := &testutil.MockUserRepo{}
	mockUser.On("GetById", mock.Anything, adminID).Return(testutil.CreateTestAdminWithID(adminID), nil)
	mockUser.On("List", mock.Anything).Return(testutil.CreateTestUserList(1, 2), nil)

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), adminID))
	rr := httptest.NewRecorder()

	h.getAll(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestUserHandler_ListUsers_Unauthorized(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	h.getAll(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestUserHandler_GetUserByID_Success(t *testing.T) {
	userID := int64(201)
	mockUser := &testutil.MockUserRepo{}
	mockUser.On("GetById", mock.Anything, userID).Return(testutil.CreateTestUserWithID(userID), nil)

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/users?id=201", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rr := httptest.NewRecorder()

	h.getById(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestUserHandler_GetUserByID_NotFound(t *testing.T) {
	userID := int64(202)
	mockUser := &testutil.MockUserRepo{}
	mockUser.On("GetById", mock.Anything, userID).Return(nil, db.ErrNotFound)

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/users?id=202", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rr := httptest.NewRecorder()

	h.getById(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestUserHandler_SearchUsers_Success(t *testing.T) {
	userID := int64(203)
	mockUser := &testutil.MockUserRepo{}
	mockUser.On("Search", mock.Anything, userID, "ali").Return(testutil.CreateTestUserList(10, 11), nil)

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/users/search?q=ali", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rr := httptest.NewRecorder()

	h.searchAll(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_Success(t *testing.T) {
	userID := int64(204)
	mockUser := &testutil.MockUserRepo{}
	existing := testutil.CreateTestUserWithID(userID)
	mockUser.On("GetById", mock.Anything, userID).Return(existing, nil)
	mockUser.On("UpdateById", mock.Anything, userID, mock.AnythingOfType("*entity.User")).Return(nil)

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	body := `{"username":"updated-user","email":"updated-user@example.com"}`
	req := httptest.NewRequest(http.MethodPut, "/users?id=204", bytes.NewBufferString(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.updateById(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_Unauthorized(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodPut, "/users?id=204", bytes.NewBufferString(`{"username":"updated"}`))
	rr := httptest.NewRecorder()

	h.updateById(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestUserHandler_DeleteUser_Success(t *testing.T) {
	userID := int64(205)
	mockUser := &testutil.MockUserRepo{}
	mockUser.On("DeleteById", mock.Anything, userID).Return(nil)

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/users?id=205", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rr := httptest.NewRecorder()

	h.deleteById(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_Unauthorized(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/users?id=205", nil)
	rr := httptest.NewRecorder()

	h.deleteById(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestUserHandler_CreateUser_Success(t *testing.T) {
	adminID := int64(206)
	mockUser := &testutil.MockUserRepo{}
	mockUser.On("GetById", mock.Anything, adminID).Return(testutil.CreateTestAdminWithID(adminID), nil)
	mockUser.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(501), nil)

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	body := `{"username":"new-user","email":"new-user@example.com","profile_pic":"avatar.png"}`
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), adminID))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.create(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestUserHandler_SearchUsers_Unauthorized(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/users/search?q=ali", nil)
	rr := httptest.NewRecorder()

	h.searchAll(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestUserHandler_SearchUsers_Error(t *testing.T) {
	userID := int64(207)
	mockUser := &testutil.MockUserRepo{}
	mockUser.On("Search", mock.Anything, userID, "ali").Return(nil, db.ErrNotFound)

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/users/search?q=ali", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rr := httptest.NewRecorder()

	h.searchAll(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestUserHandler_UpdateUser_BadID(t *testing.T) {
	userID := int64(208)
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodPut, "/users?id=bad", bytes.NewBufferString(`{"username":"updated"}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rr := httptest.NewRecorder()

	h.updateById(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_UpdateUser_InvalidJSON(t *testing.T) {
	userID := int64(209)
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodPut, "/users?id=209", bytes.NewBufferString(`{"username":`))
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rr := httptest.NewRecorder()

	h.updateById(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_UpdateUser_Error(t *testing.T) {
	userID := int64(210)
	mockUser := &testutil.MockUserRepo{}
	existing := testutil.CreateTestUserWithID(userID)
	mockUser.On("GetById", mock.Anything, userID).Return(existing, nil)
	mockUser.On("UpdateById", mock.Anything, userID, mock.AnythingOfType("*entity.User")).Return(db.ErrNotFound)

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodPut, "/users?id=210", bytes.NewBufferString(`{"username":"updated","email":"updated@example.com"}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rr := httptest.NewRecorder()

	h.updateById(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestUserHandler_DeleteUser_BadID(t *testing.T) {
	userID := int64(211)
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/users?id=bad", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rr := httptest.NewRecorder()

	h.deleteById(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_CreateUser_Unauthorized(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"username":"new-user"}`))
	rr := httptest.NewRecorder()

	h.create(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestUserHandler_CreateUser_InvalidJSON(t *testing.T) {
	adminID := int64(212)
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"username":`))
	req = req.WithContext(middleware.WithUserID(req.Context(), adminID))
	rr := httptest.NewRecorder()

	h.create(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_CreateUser_Forbidden(t *testing.T) {
	currentUserID := int64(213)
	mockUser := &testutil.MockUserRepo{}
	normalUser := testutil.CreateTestUserWithID(currentUserID)
	normalUser.Role = entity.RoleUser
	mockUser.On("GetById", mock.Anything, currentUserID).Return(normalUser, nil)

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"username":"new-user","email":"new-user@example.com"}`))
	req = req.WithContext(middleware.WithUserID(req.Context(), currentUserID))
	rr := httptest.NewRecorder()

	h.create(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestUserHandler_GetUserByID_Unauthorized(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/users?id=100", nil)
	rr := httptest.NewRecorder()

	h.getById(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestUserHandler_GetUserByID_BadID(t *testing.T) {
	userID := int64(214)
	mockUser := &testutil.MockUserRepo{}

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/users?id=bad", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rr := httptest.NewRecorder()

	h.getById(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_ListUsers_Forbidden(t *testing.T) {
	userID := int64(215)
	mockUser := &testutil.MockUserRepo{}
	normalUser := testutil.CreateTestUserWithID(userID)
	normalUser.Role = entity.RoleUser
	mockUser.On("GetById", mock.Anything, userID).Return(normalUser, nil)

	uc := usecase.NewUserUseCase(mockUser)
	h := NewUserHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), userID))
	rr := httptest.NewRecorder()

	h.getAll(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code)
	mockUser.AssertExpectations(t)
}

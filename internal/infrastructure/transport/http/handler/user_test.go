package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newUserHandler(userRepo *testutil.MockUserRepo) *UserHandler {
	uc := usecase.NewUserUseCase(userRepo)
	return NewUserHandler(uc)
}

func TestUserHandler_GetAll_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	users := []*entity.User{testutil.CreateTestUserWithID(1), testutil.CreateTestUserWithID(2)}
	userRepo.On("List", mock.Anything).Return(users, nil)

	h := newUserHandler(userRepo)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result []*entity.User
	err := json.NewDecoder(rr.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestUserHandler_GetAll_Error(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	userRepo.On("List", mock.Anything).Return(nil, errors.New("db error"))

	h := newUserHandler(userRepo)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUserHandler_GetById_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	userRepo.On("GetById", mock.Anything, int64(1)).Return(testutil.CreateTestUserWithID(1), nil)

	h := newUserHandler(userRepo)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/1/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUserHandler_GetById_BadID(t *testing.T) {
	h := newUserHandler(new(testutil.MockUserRepo))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/abc/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_GetById_NotFound(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	userRepo.On("GetById", mock.Anything, int64(999)).Return(nil, errors.New("not found"))

	h := newUserHandler(userRepo)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/999/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUserHandler_Create_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(1), nil)

	h := newUserHandler(userRepo)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"username":"newuser","email":"new@test.com","password":"pass123"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	userRepo.AssertExpectations(t)
}

func TestUserHandler_Create_BadJSON(t *testing.T) {
	h := newUserHandler(new(testutil.MockUserRepo))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("POST", "/", strings.NewReader("bad"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_Create_Error(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(0), errors.New("create failed"))

	h := newUserHandler(userRepo)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"username":"newuser","email":"new@test.com","password":"pass123"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestUserHandler_UpdateById_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	userRepo.On("UpdateById", mock.Anything, int64(1), mock.AnythingOfType("*entity.User")).Return(nil)

	h := newUserHandler(userRepo)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"username":"updated","email":"up@test.com","password":"pass123"}`
	req := httptest.NewRequest("PUT", "/1/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUserHandler_UpdateById_BadID(t *testing.T) {
	h := newUserHandler(new(testutil.MockUserRepo))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("PUT", "/bad/", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_DeleteById_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	userRepo.On("DeleteById", mock.Anything, int64(1)).Return(nil)

	h := newUserHandler(userRepo)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("DELETE", "/1/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	userRepo.AssertExpectations(t)
}

func TestUserHandler_DeleteById_BadID(t *testing.T) {
	h := newUserHandler(new(testutil.MockUserRepo))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("DELETE", "/abc/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUserHandler_SearchAll_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	users := []*entity.User{testutil.CreateTestUserWithID(2)}
	userRepo.On("Search", mock.Anything, int64(1), "test").Return(users, nil)

	h := newUserHandler(userRepo)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/search?q=test", nil)
	req = req.WithContext(middleware.ContextWithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUserHandler_SearchAll_NoAuth(t *testing.T) {
	h := newUserHandler(new(testutil.MockUserRepo))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/search?q=test", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

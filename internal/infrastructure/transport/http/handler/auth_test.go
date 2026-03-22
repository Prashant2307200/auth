package handler

import (
	"database/sql"
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
	"github.com/Prashant2307200/auth-service/pkg/hash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newAuthHandler(userRepo *testutil.MockUserRepo, businessRepo *testutil.MockBusinessRepo, tokenService *testutil.MockTokenService, cloudService *testutil.MockCloudService) *AuthHandler {
	uc := usecase.NewAuthUseCase(userRepo, businessRepo, tokenService, cloudService)
	return NewAuthHandler(uc, "test")
}

func TestAuthHandler_Register_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	businessRepo := new(testutil.MockBusinessRepo)
	tokenService := new(testutil.MockTokenService)
	cloudService := new(testutil.MockCloudService)

	userRepo.On("GetByEmail", mock.Anything, "new@example.com").Return(nil, sql.ErrNoRows)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(1), nil)
	businessRepo.On("FindAutoJoinBusinessByEmailDomain", mock.Anything, "example.com").Return(nil, nil)
	tokenService.On("GenerateRefreshToken", int64(1)).Return("refresh", nil)
	tokenService.On("StoreRefreshToken", mock.Anything, int64(1), "refresh").Return(nil)
	tokenService.On("GenerateAccessToken", int64(1)).Return("access", nil)

	h := newAuthHandler(userRepo, businessRepo, tokenService, cloudService)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"username":"newuser","email":"new@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/register/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	userRepo.AssertExpectations(t)
	tokenService.AssertExpectations(t)
}

func TestAuthHandler_Register_BadJSON(t *testing.T) {
	h := newAuthHandler(new(testutil.MockUserRepo), new(testutil.MockBusinessRepo), new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("POST", "/register/", strings.NewReader("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	h := newAuthHandler(new(testutil.MockUserRepo), new(testutil.MockBusinessRepo), new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Missing required fields
	body := `{"username":"ab","email":"bad","password":"12"}`
	req := httptest.NewRequest("POST", "/register/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	userRepo.On("GetByEmail", mock.Anything, "dup@example.com").Return(testutil.CreateTestUser(), nil)

	h := newAuthHandler(userRepo, new(testutil.MockBusinessRepo), new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"username":"newuser","email":"dup@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/register/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	tokenService := new(testutil.MockTokenService)

	hashed, err := hash.HashPassword("password123")
	assert.NoError(t, err)

	user := testutil.CreateTestUser()
	user.Password = hashed
	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
	tokenService.On("GenerateRefreshToken", int64(1)).Return("refresh", nil)
	tokenService.On("StoreRefreshToken", mock.Anything, int64(1), "refresh").Return(nil)
	tokenService.On("GenerateAccessToken", int64(1)).Return("access", nil)

	h := newAuthHandler(userRepo, nil, tokenService, new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/login/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	// Verify cookies are set
	cookies := rr.Result().Cookies()
	cookieNames := make([]string, len(cookies))
	for i, c := range cookies {
		cookieNames[i] = c.Name
	}
	assert.Contains(t, cookieNames, "access_token")
	assert.Contains(t, cookieNames, "refresh_token")
}

func TestAuthHandler_Login_BadJSON(t *testing.T) {
	h := newAuthHandler(new(testutil.MockUserRepo), nil, new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("POST", "/login/", strings.NewReader("bad"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	user := testutil.CreateTestUser()
	user.Password = "not-a-bcrypt-hash"
	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)

	h := newAuthHandler(userRepo, nil, new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"email":"test@example.com","password":"wrong"}`
	req := httptest.NewRequest("POST", "/login/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_Login_UserNotFound(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	userRepo.On("GetByEmail", mock.Anything, "noone@test.com").Return(nil, sql.ErrNoRows)

	h := newAuthHandler(userRepo, nil, new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	body := `{"email":"noone@test.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/login/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	tokenService := new(testutil.MockTokenService)
	tokenService.On("RemoveRefreshToken", mock.Anything, int64(1)).Return(nil)

	h := newAuthHandler(new(testutil.MockUserRepo), nil, tokenService, new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("DELETE", "/logout/", nil)
	req = req.WithContext(middleware.ContextWithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	tokenService.AssertExpectations(t)
}

func TestAuthHandler_Logout_NoAuth(t *testing.T) {
	h := newAuthHandler(new(testutil.MockUserRepo), nil, new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("DELETE", "/logout/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_Profile_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	user := testutil.CreateTestUserWithID(1)
	userRepo.On("GetById", mock.Anything, int64(1)).Return(user, nil)

	h := newAuthHandler(userRepo, nil, new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/profile/", nil)
	req = req.WithContext(middleware.ContextWithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result entity.User
	err := json.NewDecoder(rr.Body).Decode(&result)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
}

func TestAuthHandler_Profile_NoAuth(t *testing.T) {
	h := newAuthHandler(new(testutil.MockUserRepo), nil, new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/profile/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_DeleteProfile_Success(t *testing.T) {
	userRepo := new(testutil.MockUserRepo)
	userRepo.On("DeleteById", mock.Anything, int64(1)).Return(nil)

	h := newAuthHandler(userRepo, nil, new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("DELETE", "/profile/", nil)
	req = req.WithContext(middleware.ContextWithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	userRepo.AssertExpectations(t)
}

func TestAuthHandler_DeleteProfile_NoAuth(t *testing.T) {
	h := newAuthHandler(new(testutil.MockUserRepo), nil, new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("DELETE", "/profile/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	tokenService := new(testutil.MockTokenService)
	tokenService.On("VerifyRefreshToken", mock.Anything, "old-refresh").Return("1", nil)
	tokenService.On("GetRefreshToken", mock.Anything, int64(1)).Return("old-refresh", nil)
	tokenService.On("GenerateRefreshToken", int64(1)).Return("new-refresh", nil)
	tokenService.On("GenerateAccessToken", int64(1)).Return("new-access", nil)
	tokenService.On("StoreRefreshToken", mock.Anything, int64(1), "new-refresh").Return(nil)

	h := newAuthHandler(new(testutil.MockUserRepo), nil, tokenService, new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/refresh/", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh"})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	tokenService.AssertExpectations(t)
}

func TestAuthHandler_Refresh_NoCookie(t *testing.T) {
	h := newAuthHandler(new(testutil.MockUserRepo), nil, new(testutil.MockTokenService), new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/refresh/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	tokenService := new(testutil.MockTokenService)
	tokenService.On("VerifyRefreshToken", mock.Anything, "bad-token").Return("", errors.New("invalid"))

	h := newAuthHandler(new(testutil.MockUserRepo), nil, tokenService, new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/refresh/", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "bad-token"})
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthHandler_PublicKey_Success(t *testing.T) {
	tokenService := new(testutil.MockTokenService)
	tokenService.On("GetPublicKeyPEM").Return([]byte("-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----"), nil)

	h := newAuthHandler(new(testutil.MockUserRepo), nil, tokenService, new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/public-key", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/x-pem-file", rr.Header().Get("Content-Type"))
}

func TestAuthHandler_PublicKey_Error(t *testing.T) {
	tokenService := new(testutil.MockTokenService)
	tokenService.On("GetPublicKeyPEM").Return(nil, errors.New("key not found"))

	h := newAuthHandler(new(testutil.MockUserRepo), nil, tokenService, new(testutil.MockCloudService))
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/public-key", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

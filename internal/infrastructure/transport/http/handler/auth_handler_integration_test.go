package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	uutils "github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/Prashant2307200/auth-service/pkg/hash"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_Register_Success(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	mockUser.On("GetByEmail", mock.Anything, "new-user@example.com").Return(nil, nil)
	mockUser.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(int64(1), nil)
	mockBusiness.On("FindAutoJoinBusinessByEmailDomain", mock.Anything, "example.com").Return(nil, nil)
	mockBusiness.On("CreateWithOwner", mock.Anything, mock.AnythingOfType("*entity.Business"), int64(1)).Return(int64(99), nil)
	mockToken.On("GenerateRefreshToken", int64(1)).Return("refresh-token", nil)
	mockToken.On("StoreRefreshToken", mock.Anything, int64(1), "refresh-token").Return(nil)
	mockToken.On("GenerateAccessToken", int64(1), int64(99)).Return("access-token", nil)

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	body := `{"username":"newuser","email":"new-user@example.com","password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/register/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.register(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.NotEmpty(t, rr.Result().Cookies())

	mockUser.AssertExpectations(t)
	mockBusiness.AssertExpectations(t)
	mockToken.AssertExpectations(t)
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	mockUser.On("GetByEmail", mock.Anything, "existing@example.com").Return(testutil.CreateTestUserWithEmail("existing@example.com"), nil)

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	body := `{"username":"existing","email":"existing@example.com","password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/register/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.register(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	var er uutils.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, uutils.INTERNAL_ERROR, er.Code)

	mockUser.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	hashed, err := hash.HashPassword("Password123!")
	require.NoError(t, err)

	user := &entity.User{ID: 1, Email: "login@example.com", Password: hashed}
	mockUser.On("GetByEmail", mock.Anything, "login@example.com").Return(user, nil)
	mockToken.On("GenerateRefreshToken", int64(1)).Return("refresh-token", nil)
	mockToken.On("StoreRefreshToken", mock.Anything, int64(1), "refresh-token").Return(nil)
	mockToken.On("GenerateAccessToken", int64(1)).Return("access-token", nil)

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	body := `{"email":"login@example.com","password":"Password123!"}`
	req := httptest.NewRequest(http.MethodPost, "/login/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.login(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	cookies := rr.Result().Cookies()
	require.Len(t, cookies, 2)
	require.Equal(t, "access_token", cookies[0].Name)
	require.Equal(t, "refresh_token", cookies[1].Name)

	mockUser.AssertExpectations(t)
	mockToken.AssertExpectations(t)
}

func TestAuthHandler_Login_WrongPassword(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	hashed, err := hash.HashPassword("Password123!")
	require.NoError(t, err)

	user := &entity.User{ID: 1, Email: "login@example.com", Password: hashed}
	mockUser.On("GetByEmail", mock.Anything, "login@example.com").Return(user, nil)

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	body := `{"email":"login@example.com","password":"WrongPassword"}`
	req := httptest.NewRequest(http.MethodPost, "/login/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.login(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)

	mockUser.AssertExpectations(t)
}

func TestAuthHandler_Profile_Success(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	expectedUser := testutil.CreateTestUserWithID(42)
	mockUser.On("GetById", mock.Anything, int64(42)).Return(expectedUser, nil)

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodGet, "/profile/", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 42))
	rr := httptest.NewRecorder()
	h.profile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got entity.User
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Equal(t, int64(42), got.ID)

	mockUser.AssertExpectations(t)
}

func TestAuthHandler_DeleteProfile_Unauthorized(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodDelete, "/profile/", nil)
	rr := httptest.NewRecorder()
	h.deleteProfile(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	var er uutils.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, uutils.UNAUTHORIZED, er.Code)
}

func TestAuthHandler_DeleteProfile_Success(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	mockUser.On("DeleteById", mock.Anything, int64(12)).Return(nil)

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodDelete, "/profile/", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 12))
	rr := httptest.NewRecorder()
	h.deleteProfile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestAuthHandler_Refresh_Success(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	mockToken.On("VerifyRefreshToken", mock.Anything, "old-refresh").Return("7", nil)
	mockToken.On("GetRefreshToken", mock.Anything, int64(7)).Return("old-refresh", nil)
	mockToken.On("GenerateRefreshToken", int64(7)).Return("new-refresh", nil)
	mockToken.On("GenerateAccessToken", int64(7)).Return("new-access", nil)
	mockToken.On("StoreRefreshToken", mock.Anything, int64(7), "new-refresh").Return(nil)

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodGet, "/refresh/", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh"})
	rr := httptest.NewRecorder()
	h.refresh(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	cookies := rr.Result().Cookies()
	require.Len(t, cookies, 2)
	require.Equal(t, "access_token", cookies[0].Name)
	require.Equal(t, "refresh_token", cookies[1].Name)

	mockToken.AssertExpectations(t)
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	mockToken.On("VerifyRefreshToken", mock.Anything, "bad-refresh").Return("", errors.New("invalid token"))

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodGet, "/refresh/", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "bad-refresh"})
	rr := httptest.NewRecorder()
	h.refresh(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	mockToken.AssertExpectations(t)
}

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	mockToken.On("RemoveRefreshToken", mock.Anything, int64(42)).Return(nil)

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodDelete, "/logout/", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 42))
	rr := httptest.NewRecorder()
	h.logout(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockToken.AssertExpectations(t)
}

func TestAuthHandler_UpdateProfile_Unauthorized(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodPut, "/profile/", bytes.NewReader([]byte(`{"username":"newname"}`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.updateProfile(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	var er uutils.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, uutils.UNAUTHORIZED, er.Code)
}

func TestAuthHandler_PublicKey_Error(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}

	mockToken.On("GetPublicKeyPEM").Return(nil, errors.New("key unavailable"))

	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodGet, "/public-key", nil)
	rr := httptest.NewRecorder()
	h.publicKey(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	mockToken.AssertExpectations(t)
}

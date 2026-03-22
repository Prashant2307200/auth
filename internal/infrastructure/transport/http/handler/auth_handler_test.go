package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	uutils "github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils"
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthHandler_Register_MalformedJSON(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}
	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodPost, "/register/", bytes.NewReader([]byte("invalid json")))
	rr := httptest.NewRecorder()
	h.register(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	var er uutils.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, uutils.BAD_REQUEST, er.Code)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}
	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	body := `{"username":"ab","email":"x","password":"short"}` // username too short, invalid email
	req := httptest.NewRequest(http.MethodPost, "/register/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.register(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	_, ok := got["errors"]
	require.True(t, ok)
}

func TestAuthHandler_Login_MalformedJSON(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}
	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodPost, "/login/", bytes.NewReader([]byte("{")))
	rr := httptest.NewRecorder()
	h.login(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	var er uutils.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, uutils.BAD_REQUEST, er.Code)
}

func TestAuthHandler_Profile_Unauthorized(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}
	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodGet, "/profile/", nil)
	rr := httptest.NewRecorder()
	h.profile(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	var er uutils.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, uutils.UNAUTHORIZED, er.Code)
}

func TestAuthHandler_PublicKey_Success(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}
	mockToken.On("GetPublicKeyPEM").Return([]byte("-----BEGIN PUBLIC KEY-----\nfake\n-----END PUBLIC KEY-----"), nil)
	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodGet, "/public-key", nil)
	rr := httptest.NewRecorder()
	h.publicKey(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Header().Get("Content-Type"), "pem")
	mockToken.AssertExpectations(t)
}

func TestAuthHandler_Refresh_NoCookie(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}
	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodGet, "/refresh/", nil)
	rr := httptest.NewRecorder()
	h.refresh(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	var er uutils.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, uutils.UNAUTHORIZED, er.Code)
}

func TestAuthHandler_Logout_Unauthorized(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}
	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodDelete, "/logout/", nil)
	rr := httptest.NewRecorder()
	h.logout(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	var er uutils.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, uutils.UNAUTHORIZED, er.Code)
}

func TestAuthHandler_UploadSignature_Unauthorized(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}
	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodGet, "/upload-signature", nil)
	rr := httptest.NewRecorder()
	h.uploadSignature(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	var er uutils.ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&er))
	require.Equal(t, uutils.UNAUTHORIZED, er.Code)
}

func TestAuthHandler_UploadSignature_Success(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}
	sig := &interfaces.UploadSignature{
		UploadURL: "https://api.cloudinary.com/v1_1/demo/image/upload",
		APIKey:    "key",
		Signature: "sig",
		Timestamp: "123",
		Folder:    "profile_pics",
		PublicID:  "profile_1_123",
		CloudName: "demo",
	}
	mockCloud.On("GenerateUploadSignature", mock.Anything, int64(42)).Return(sig, nil)
	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	req := httptest.NewRequest(http.MethodGet, "/upload-signature", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 42))
	rr := httptest.NewRecorder()
	h.uploadSignature(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Equal(t, "https://api.cloudinary.com/v1_1/demo/image/upload", got["upload_url"])
	require.Equal(t, "key", got["api_key"])
	mockCloud.AssertExpectations(t)
}

func TestAuthHandler_UpdateProfile_JSONWithProfilePic(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}
	mockUser.On("GetById", mock.Anything, int64(1)).Return(testutil.CreateTestUserWithID(1), nil)
	mockUser.On("UpdateById", mock.Anything, int64(1), mock.MatchedBy(func(u *entity.User) bool {
		return u.ProfilePic == "https://res.cloudinary.com/demo/image/upload/v1/profile_pics/profile_1_123.jpg"
	})).Return(nil)
	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	body := `{"profile_pic":"https://res.cloudinary.com/demo/image/upload/v1/profile_pics/profile_1_123.jpg"}`
	req := httptest.NewRequest(http.MethodPut, "/profile/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.updateProfile(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockUser.AssertExpectations(t)
}

func TestAuthHandler_UpdateProfile_InvalidProfilePic(t *testing.T) {
	mockUser := &testutil.MockUserRepo{}
	mockBusiness := &testutil.MockBusinessRepo{}
	mockToken := &testutil.MockTokenService{}
	mockCloud := &testutil.MockCloudService{}
	mockUser.On("GetById", mock.Anything, int64(1)).Return(testutil.CreateTestUserWithID(1), nil)
	uc := usecase.NewAuthUseCase(mockUser, mockBusiness, mockToken, mockCloud)
	h := NewAuthHandler(uc, "dev")

	body := `{"profile_pic":"javascript:alert(1)"}`
	req := httptest.NewRequest(http.MethodPut, "/profile/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.updateProfile(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	_, ok := got["errors"]
	require.True(t, ok)
	mockUser.AssertExpectations(t)
}

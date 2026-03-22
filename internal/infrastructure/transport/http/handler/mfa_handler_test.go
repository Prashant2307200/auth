package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockMFAUsecase struct {
	mock.Mock
}

func (m *mockMFAUsecase) Setup(ctx context.Context, userID int64, email string) (*usecase.MFASetupResult, error) {
	args := m.Called(ctx, userID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.MFASetupResult), args.Error(1)
}

func (m *mockMFAUsecase) Enable(ctx context.Context, userID int64, code string) ([]string, error) {
	args := m.Called(ctx, userID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockMFAUsecase) Disable(ctx context.Context, userID int64, code string) error {
	args := m.Called(ctx, userID, code)
	return args.Error(0)
}

func (m *mockMFAUsecase) Verify(ctx context.Context, userID int64, code string) error {
	args := m.Called(ctx, userID, code)
	return args.Error(0)
}

func (m *mockMFAUsecase) VerifyBackupCode(ctx context.Context, userID int64, code string) error {
	args := m.Called(ctx, userID, code)
	return args.Error(0)
}

func (m *mockMFAUsecase) RegenerateBackupCodes(ctx context.Context, userID int64, code string) ([]string, error) {
	args := m.Called(ctx, userID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockMFAUsecase) IsEnabled(ctx context.Context, userID int64) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

type mockUserRepoForMFA struct {
	mock.Mock
}

func (m *mockUserRepoForMFA) List(ctx context.Context) ([]*entity.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *mockUserRepoForMFA) GetById(ctx context.Context, id int64) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepoForMFA) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepoForMFA) GetByGoogleID(ctx context.Context, googleID string) (*entity.User, error) {
	args := m.Called(ctx, googleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepoForMFA) Create(ctx context.Context, user *entity.User) (int64, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockUserRepoForMFA) UpdateById(ctx context.Context, id int64, user *entity.User) error {
	args := m.Called(ctx, id, user)
	return args.Error(0)
}

func (m *mockUserRepoForMFA) UpdatePassword(ctx context.Context, id int64, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *mockUserRepoForMFA) DeleteById(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepoForMFA) MarkEmailVerified(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockUserRepoForMFA) LinkGoogleID(ctx context.Context, userID int64, googleID string) error {
	args := m.Called(ctx, userID, googleID)
	return args.Error(0)
}

func (m *mockUserRepoForMFA) Search(ctx context.Context, currentID int64, search string) ([]*entity.User, error) {
	args := m.Called(ctx, currentID, search)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.User), args.Error(1)
}

func TestMFAHandler_Setup_Unauthorized(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}
	h := NewMFAHandler(uc, userRepo)

	req := httptest.NewRequest(http.MethodPost, "/mfa/setup", nil)
	rr := httptest.NewRecorder()
	h.setup(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMFAHandler_Setup_Success(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}

	user := &entity.User{ID: 1, Email: "user@example.com"}
	userRepo.On("GetById", mock.Anything, int64(1)).Return(user, nil)
	uc.On("Setup", mock.Anything, int64(1), "user@example.com").Return(&usecase.MFASetupResult{
		Secret:    "SECRETBASE32",
		QRCodeURI: "otpauth://totp/App:user@example.com?secret=SECRETBASE32",
	}, nil)

	h := NewMFAHandler(uc, userRepo)

	req := httptest.NewRequest(http.MethodPost, "/mfa/setup", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.setup(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Equal(t, "SECRETBASE32", got["secret"])
	qrURI, ok := got["qr_code_uri"].(string)
	require.True(t, ok, "qr_code_uri should be a string")
	require.Contains(t, qrURI, "otpauth://")
	uc.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestMFAHandler_Setup_AlreadyEnabled(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}

	user := &entity.User{ID: 1, Email: "user@example.com"}
	userRepo.On("GetById", mock.Anything, int64(1)).Return(user, nil)
	uc.On("Setup", mock.Anything, int64(1), "user@example.com").Return(nil, usecase.ErrMFAAlreadyEnabled)

	h := NewMFAHandler(uc, userRepo)

	req := httptest.NewRequest(http.MethodPost, "/mfa/setup", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.setup(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	uc.AssertExpectations(t)
}

func TestMFAHandler_Enable_Unauthorized(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}
	h := NewMFAHandler(uc, userRepo)

	body := `{"code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/mfa/enable", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.enable(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMFAHandler_Enable_MalformedJSON(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}
	h := NewMFAHandler(uc, userRepo)

	req := httptest.NewRequest(http.MethodPost, "/mfa/enable", bytes.NewReader([]byte("{")))
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.enable(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestMFAHandler_Enable_InvalidCode(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}
	uc.On("Enable", mock.Anything, int64(1), "000000").Return(nil, usecase.ErrInvalidTOTPCode)
	h := NewMFAHandler(uc, userRepo)

	body := `{"code":"000000"}`
	req := httptest.NewRequest(http.MethodPost, "/mfa/enable", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.enable(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	uc.AssertExpectations(t)
}

func TestMFAHandler_Enable_Success(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}
	backupCodes := []string{"code1", "code2", "code3"}
	uc.On("Enable", mock.Anything, int64(1), "123456").Return(backupCodes, nil)
	h := NewMFAHandler(uc, userRepo)

	body := `{"code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/mfa/enable", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.enable(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Equal(t, "MFA enabled successfully", got["message"])
	require.NotNil(t, got["backup_codes"])
	uc.AssertExpectations(t)
}

func TestMFAHandler_Verify_Unauthorized(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}
	h := NewMFAHandler(uc, userRepo)

	body := `{"code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/mfa/verify", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.verify(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMFAHandler_Verify_Success(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}
	uc.On("Verify", mock.Anything, int64(1), "123456").Return(nil)
	h := NewMFAHandler(uc, userRepo)

	body := `{"code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/mfa/verify", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.verify(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	uc.AssertExpectations(t)
}

func TestMFAHandler_Disable_Unauthorized(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}
	h := NewMFAHandler(uc, userRepo)

	body := `{"code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/mfa/disable", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.disable(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMFAHandler_Disable_Success(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}
	uc.On("Disable", mock.Anything, int64(1), "123456").Return(nil)
	h := NewMFAHandler(uc, userRepo)

	body := `{"code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/mfa/disable", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.disable(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	uc.AssertExpectations(t)
}

func TestMFAHandler_Status_Unauthorized(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}
	h := NewMFAHandler(uc, userRepo)

	req := httptest.NewRequest(http.MethodGet, "/mfa/status", nil)
	rr := httptest.NewRecorder()
	h.status(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMFAHandler_Status_Success(t *testing.T) {
	uc := &mockMFAUsecase{}
	userRepo := &mockUserRepoForMFA{}
	uc.On("IsEnabled", mock.Anything, int64(1)).Return(true, nil)
	h := NewMFAHandler(uc, userRepo)

	req := httptest.NewRequest(http.MethodGet, "/mfa/status", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.status(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Equal(t, true, got["mfa_enabled"])
	uc.AssertExpectations(t)
}

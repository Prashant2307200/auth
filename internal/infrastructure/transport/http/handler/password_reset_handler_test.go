package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/usecase"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockPasswordResetUsecase struct {
	mock.Mock
}

func (m *mockPasswordResetUsecase) RequestReset(ctx context.Context, email string) (string, error) {
	args := m.Called(ctx, email)
	return args.String(0), args.Error(1)
}

func (m *mockPasswordResetUsecase) ResetPassword(ctx context.Context, token, newPassword string) error {
	args := m.Called(ctx, token, newPassword)
	return args.Error(0)
}

func TestPasswordResetHandler_ForgotPassword_MalformedJSON(t *testing.T) {
	uc := &mockPasswordResetUsecase{}
	h := NewPasswordResetHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader([]byte("invalid")))
	rr := httptest.NewRecorder()
	h.forgotPassword(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestPasswordResetHandler_ForgotPassword_Success(t *testing.T) {
	uc := &mockPasswordResetUsecase{}
	uc.On("RequestReset", mock.Anything, "user@example.com").Return("token123", nil)
	h := NewPasswordResetHandler(uc)

	body := `{"email":"user@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.forgotPassword(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Equal(t, "success", got["status"])
	uc.AssertExpectations(t)
}

func TestPasswordResetHandler_ResetPassword_MalformedJSON(t *testing.T) {
	uc := &mockPasswordResetUsecase{}
	h := NewPasswordResetHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader([]byte("{")))
	rr := httptest.NewRecorder()
	h.resetPassword(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestPasswordResetHandler_ResetPassword_InvalidToken(t *testing.T) {
	uc := &mockPasswordResetUsecase{}
	uc.On("ResetPassword", mock.Anything, "invalidtoken", "newpassword123").Return(usecase.ErrTokenNotFound)
	h := NewPasswordResetHandler(uc)

	body := `{"token":"invalidtoken","new_password":"newpassword123"}`
	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.resetPassword(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	uc.AssertExpectations(t)
}

func TestPasswordResetHandler_ResetPassword_Success(t *testing.T) {
	uc := &mockPasswordResetUsecase{}
	uc.On("ResetPassword", mock.Anything, "validtoken", "newpassword123").Return(nil)
	h := NewPasswordResetHandler(uc)

	body := `{"token":"validtoken","new_password":"newpassword123"}`
	req := httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.resetPassword(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Equal(t, "success", got["status"])
	uc.AssertExpectations(t)
}

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRevokeInvitation_Handler_Success(t *testing.T) {
	ucMock := &testutil.MockTeamUsecase{}
	ucMock.On("RevokeInvitation", mock.Anything, "valid-token").Return(nil)

	handler := &TeamHandler{UC: ucMock, AMW: func(h http.Handler) http.Handler { return h }}

	req := httptest.NewRequest("POST", "/api/v1/team/invites/valid-token/revoke", nil)
	req.SetPathValue("token", "valid-token")
	w := httptest.NewRecorder()

	handler.revokeInvitation(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "invitation revoked", resp["message"])
	ucMock.AssertExpectations(t)
}

func TestRevokeInvitation_Handler_NotFound(t *testing.T) {
	ucMock := &testutil.MockTeamUsecase{}
	ucMock.On("RevokeInvitation", mock.Anything, "nonexistent").Return(testutil.ErrNotFound)

	handler := &TeamHandler{UC: ucMock, AMW: func(h http.Handler) http.Handler { return h }}

	req := httptest.NewRequest("POST", "/api/v1/team/invites/nonexistent/revoke", nil)
	req.SetPathValue("token", "nonexistent")
	w := httptest.NewRecorder()

	handler.revokeInvitation(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	ucMock.AssertExpectations(t)
}

func TestRevokeInvitation_Handler_AlreadyAccepted(t *testing.T) {
	ucMock := &testutil.MockTeamUsecase{}
	ucMock.On("RevokeInvitation", mock.Anything, "accepted-token").Return(testutil.ErrCannotRevoke)

	handler := &TeamHandler{UC: ucMock, AMW: func(h http.Handler) http.Handler { return h }}

	req := httptest.NewRequest("POST", "/api/v1/team/invites/accepted-token/revoke", nil)
	req.SetPathValue("token", "accepted-token")
	w := httptest.NewRecorder()

	handler.revokeInvitation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRevokeInvitation_Handler_InvalidToken(t *testing.T) {
	handler := &TeamHandler{UC: nil, AMW: func(h http.Handler) http.Handler { return h }}

	req := httptest.NewRequest("POST", "/api/v1/team/invites//revoke", nil)
	w := httptest.NewRecorder()

	handler.revokeInvitation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

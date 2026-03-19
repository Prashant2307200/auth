package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTeamHandler_InviteUser_InvalidEmail(t *testing.T) {
	// Setup handler with mock usecase
	handler := NewTeamHandler(nil, nil)

	req := inviteRequest{Email: "invalid-email", Role: 1}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/team/invite", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.invite(w, httpReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeamHandler_InviteUser_InvalidRole(t *testing.T) {
	handler := NewTeamHandler(nil, nil)
	req := inviteRequest{Email: "user@example.com", Role: 5}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/team/invite", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.invite(w, httpReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeamHandler_UpdateMemberRole_InvalidID(t *testing.T) {
	handler := NewTeamHandler(nil, nil)
	httpReq := httptest.NewRequest(http.MethodPatch, "/api/v1/team/members/invalid-id/role", bytes.NewReader([]byte(`{"role":2}`)))
	w := httptest.NewRecorder()
	handler.updateMemberRole(w, httpReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

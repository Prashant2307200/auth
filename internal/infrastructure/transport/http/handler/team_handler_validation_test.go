package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestTeamHandler_InviteUser_InvalidEmail(t *testing.T) {
	// Setup handler with mock usecase that provides validation
	handler := NewTeamHandler(&mockTeamUC{}, nil)

	req := inviteRequest{Email: "invalid-email", Role: 1}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/team/invite", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.invite(w, httpReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeamHandler_InviteUser_InvalidRole(t *testing.T) {
	handler := NewTeamHandler(&mockTeamUC{}, nil)
	req := inviteRequest{Email: "user@example.com", Role: 5}
	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/team/invite", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.invite(w, httpReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTeamHandler_UpdateMemberRole_InvalidID(t *testing.T) {
	handler := NewTeamHandler(&mockTeamUC{}, nil)
	httpReq := httptest.NewRequest(http.MethodPatch, "/team/members/invalid-id/role", bytes.NewReader([]byte(`{"role":2}`)))
	httpReq.SetPathValue("id", "invalid-id")
	w := httptest.NewRecorder()
	handler.updateMemberRole(w, httpReq)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// mockTeamUC is a minimal mock implementing usecase.TeamUsecase for handler validation tests
type mockTeamUC struct{}

func (m *mockTeamUC) InviteUser(ctx context.Context, businessID int64, email string, role int) (string, error) {
	return "", nil
}
func (m *mockTeamUC) AcceptInvitation(ctx context.Context, inviteToken string) error { return nil }
func (m *mockTeamUC) RevokeInvitation(ctx context.Context, inviteToken string) error { return nil }
func (m *mockTeamUC) ListMembers(ctx context.Context, businessID int64) ([]*entity.BusinessMember, error) {
	return nil, nil
}
func (m *mockTeamUC) RemoveMember(ctx context.Context, businessID int64, memberID int64) error {
	return nil
}
func (m *mockTeamUC) UpdateMemberRole(ctx context.Context, businessID int64, memberID int64, newRole int) error {
	return nil
}
func (m *mockTeamUC) ValidateInviteEmail(email string) error {
	if email == "invalid-email" {
		return errors.New("invalid email format")
	}
	return nil
}
func (m *mockTeamUC) ValidateRole(role int) error {
	if role < 1 || role > 4 {
		return errors.New("role must be between 1 and 4")
	}
	return nil
}

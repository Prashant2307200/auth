package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
)

// minimal mock implementing usecase.TeamUsecase for contract tests
type mockTeamUsecase struct{}

func (m *mockTeamUsecase) InviteUser(ctx context.Context, businessID int64, email string, role int) (string, error) {
	return "tok_123", nil
}
func (m *mockTeamUsecase) AcceptInvitation(ctx context.Context, inviteToken string) error {
	return nil
}
func (m *mockTeamUsecase) RevokeInvitation(ctx context.Context, inviteToken string) error {
	return nil
}
func (m *mockTeamUsecase) ListMembers(ctx context.Context, businessID int64) ([]*entity.BusinessMember, error) {
	return []*entity.BusinessMember{{ID: 1, BusinessID: businessID, Email: "a@a.com", RoleID: 1}}, nil
}
func (m *mockTeamUsecase) RemoveMember(ctx context.Context, businessID int64, memberID int64) error {
	return nil
}
func (m *mockTeamUsecase) UpdateMemberRole(ctx context.Context, businessID int64, memberID int64, newRole int) error {
	return nil
}
func (m *mockTeamUsecase) ValidateInviteEmail(email string) error { return nil }
func (m *mockTeamUsecase) ValidateRole(role int) error         { return nil }

func TestInviteResponseContract(t *testing.T) {
	h := NewTeamHandler(&mockTeamUsecase{}, func(h http.Handler) http.Handler { return h })

	body := strings.NewReader(`{"email":"a@a.com","role":1}`)
	req := httptest.NewRequest("POST", "/api/v1/team/invite", body)
	req = middleware.WithTenantID(req, 1)
	w := httptest.NewRecorder()
	h.invite(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected application/json content type, got %q", ct)
	}
	b, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(b), "invite_token") || !strings.Contains(string(b), "invitation sent") {
		t.Fatalf("unexpected body: %s", string(b))
	}
}

func TestUpdateMemberRoleResponseContract(t *testing.T) {
	h := NewTeamHandler(&mockTeamUsecase{}, func(h http.Handler) http.Handler { return h })

	req := httptest.NewRequest("PATCH", "/api/v1/team/members/123/role", strings.NewReader(`{"role":2}`))
	req.SetPathValue("id", "123")
	req = middleware.WithTenantID(req, 1)
	w := httptest.NewRecorder()
	h.updateMemberRole(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected application/json content type, got %q", ct)
	}
	b, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(b), "updated") {
		t.Fatalf("unexpected body: %s", string(b))
	}
}


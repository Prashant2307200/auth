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
	"github.com/Prashant2307200/auth-service/internal/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTeamHandler_Invite_Success(t *testing.T) {
	mockTeam := &testutil.MockTeamUsecase{}
	mockTeam.On("ValidateInviteEmail", "member@example.com").Return(nil)
	mockTeam.On("ValidateRole", 2).Return(nil)
	mockTeam.On("InviteUser", mock.Anything, int64(55), "member@example.com", 2).Return("invite-token", nil)

	h := NewTeamHandler(mockTeam, func(next http.Handler) http.Handler { return next })

	body := `{"email":"member@example.com","role":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/team/invite", bytes.NewReader([]byte(body)))
	req = middleware.WithTenantID(req, 55)
	rr := httptest.NewRecorder()
	h.invite(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Equal(t, "invite-token", got["invite_token"])

	mockTeam.AssertExpectations(t)
}

func TestTeamHandler_Invite_Unauthorized(t *testing.T) {
	mockTeam := &testutil.MockTeamUsecase{}
	mockTeam.On("ValidateInviteEmail", "member@example.com").Return(nil)
	mockTeam.On("ValidateRole", 2).Return(nil)

	h := NewTeamHandler(mockTeam, func(next http.Handler) http.Handler { return next })

	body := `{"email":"member@example.com","role":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/team/invite", bytes.NewReader([]byte(body)))
	rr := httptest.NewRecorder()
	h.invite(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
	mockTeam.AssertExpectations(t)
}

func TestTeamHandler_ListMembers_Success(t *testing.T) {
	mockTeam := &testutil.MockTeamUsecase{}
	members := []*entity.BusinessMember{{ID: 1, BusinessID: 88}, {ID: 2, BusinessID: 88}}
	mockTeam.On("ListMembers", mock.Anything, int64(88)).Return(members, nil)

	h := NewTeamHandler(mockTeam, func(next http.Handler) http.Handler { return next })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/team/members", nil)
	req = middleware.WithTenantID(req, 88)
	rr := httptest.NewRecorder()
	h.listMembers(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got []entity.BusinessMember
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	require.Len(t, got, 2)

	mockTeam.AssertExpectations(t)
}

func TestTeamHandler_ListMembers_Unauthorized(t *testing.T) {
	mockTeam := &testutil.MockTeamUsecase{}
	h := NewTeamHandler(mockTeam, func(next http.Handler) http.Handler { return next })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/team/members", nil)
	rr := httptest.NewRecorder()
	h.listMembers(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestTeamHandler_UpdateMemberRole_Success(t *testing.T) {
	mockTeam := &testutil.MockTeamUsecase{}
	mockTeam.On("ValidateRole", 3).Return(nil)
	mockTeam.On("UpdateMemberRole", mock.Anything, int64(66), int64(7), 3).Return(nil)

	h := NewTeamHandler(mockTeam, func(next http.Handler) http.Handler { return next })

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/team/members/7/role", bytes.NewReader([]byte(`{"role":3}`)))
	req = middleware.WithTenantID(req, 66)
	rr := httptest.NewRecorder()
	h.updateMemberRole(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockTeam.AssertExpectations(t)
}

func TestTeamHandler_RemoveMember_Success(t *testing.T) {
	mockTeam := &testutil.MockTeamUsecase{}
	mockTeam.On("RemoveMember", mock.Anything, int64(77), int64(9)).Return(nil)

	h := NewTeamHandler(mockTeam, func(next http.Handler) http.Handler { return next })

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/team/9", nil)
	req = middleware.WithTenantID(req, 77)
	rr := httptest.NewRecorder()
	h.removeMember(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	mockTeam.AssertExpectations(t)
}

func TestTeamHandler_RemoveMember_Unauthorized(t *testing.T) {
	mockTeam := &testutil.MockTeamUsecase{}
	h := NewTeamHandler(mockTeam, nil)

	authGate := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if middleware.GetTenantID(r) == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/team/9", nil)
	rr := httptest.NewRecorder()
	authGate(http.HandlerFunc(h.removeMember)).ServeHTTP(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestTeamHandler_Invite_InvalidEmail(t *testing.T) {
	mockTeam := &testutil.MockTeamUsecase{}
	mockTeam.On("ValidateInviteEmail", "bad-email").Return(errors.New("invalid email"))

	h := NewTeamHandler(mockTeam, func(next http.Handler) http.Handler { return next })

	body := `{"email":"bad-email","role":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/team/invite", bytes.NewReader([]byte(body)))
	req = middleware.WithTenantID(req, 55)
	rr := httptest.NewRecorder()
	h.invite(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	mockTeam.AssertExpectations(t)
}

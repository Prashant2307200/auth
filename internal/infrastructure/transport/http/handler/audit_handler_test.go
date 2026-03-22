package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAuditRepoForHandler struct {
	mock.Mock
}

func (m *mockAuditRepoForHandler) Log(ctx context.Context, log *entity.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *mockAuditRepoForHandler) GetByID(ctx context.Context, id int64) (*entity.AuditLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.AuditLog), args.Error(1)
}

func (m *mockAuditRepoForHandler) ListByBusiness(ctx context.Context, businessID int64, limit, offset int) ([]*entity.AuditLog, error) {
	args := m.Called(ctx, businessID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditLog), args.Error(1)
}

func (m *mockAuditRepoForHandler) ListByUser(ctx context.Context, businessID, userID int64, limit, offset int) ([]*entity.AuditLog, error) {
	args := m.Called(ctx, businessID, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditLog), args.Error(1)
}

func (m *mockAuditRepoForHandler) ListWithFilter(ctx context.Context, businessID int64, userID *int64, action, fromTime, toTime string, limit, offset int) ([]*entity.AuditLog, error) {
	args := m.Called(ctx, businessID, userID, action, fromTime, toTime, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditLog), args.Error(1)
}

func (m *mockAuditRepoForHandler) Export(ctx context.Context, businessID int64) ([]*entity.AuditLog, error) {
	args := m.Called(ctx, businessID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.AuditLog), args.Error(1)
}

func TestAuditHandler_ListAuditLogs_Unauthorized(t *testing.T) {
	repo := &mockAuditRepoForHandler{}
	h := NewAuditHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	rr := httptest.NewRecorder()
	h.listAuditLogs(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuditHandler_ListAuditLogs_MissingTenantID(t *testing.T) {
	repo := &mockAuditRepoForHandler{}
	h := NewAuditHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), 1))
	rr := httptest.NewRecorder()
	h.listAuditLogs(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAuditHandler_ListAuditLogs_Success(t *testing.T) {
	repo := &mockAuditRepoForHandler{}
	logs := []*entity.AuditLog{
		{
			ID:         1,
			BusinessID: 100,
			UserID:     1,
			Action:     entity.AuditActionUserLogin,
			IPAddress:  "192.168.1.1",
			UserAgent:  "Mozilla/5.0",
			CreatedAt:  time.Now(),
		},
		{
			ID:         2,
			BusinessID: 100,
			UserID:     2,
			Action:     entity.AuditActionUserMFAEnabled,
			IPAddress:  "192.168.1.2",
			UserAgent:  "Chrome",
			CreatedAt:  time.Now(),
		},
	}
	repo.On("ListWithFilter", mock.Anything, int64(100), (*int64)(nil), "", "", "", 20, 0).Return(logs, nil)
	h := NewAuditHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	ctx := middleware.WithUserID(req.Context(), 1)
	req = middleware.WithTenantID(req.WithContext(ctx), 100)
	rr := httptest.NewRecorder()
	h.listAuditLogs(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	auditLogs := got["audit_logs"].([]any)
	require.Len(t, auditLogs, 2)
	repo.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogs_WithFilters(t *testing.T) {
	repo := &mockAuditRepoForHandler{}
	userID := int64(5)
	logs := []*entity.AuditLog{
		{
			ID:         3,
			BusinessID: 100,
			UserID:     userID,
			Action:     entity.AuditActionUserLogin,
			IPAddress:  "10.0.0.1",
			CreatedAt:  time.Now(),
		},
	}
	repo.On("ListWithFilter", mock.Anything, int64(100), &userID, "user.login", "2025-01-01T00:00:00Z", "2025-12-31T23:59:59Z", 10, 10).Return(logs, nil)
	h := NewAuditHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/audit-logs?user_id=5&action=user.login&from=2025-01-01T00:00:00Z&to=2025-12-31T23:59:59Z&limit=10&page=2", nil)
	ctx := middleware.WithUserID(req.Context(), 1)
	req = middleware.WithTenantID(req.WithContext(ctx), 100)
	rr := httptest.NewRecorder()
	h.listAuditLogs(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	auditLogs := got["audit_logs"].([]any)
	require.Len(t, auditLogs, 1)
	repo.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogs_EmptyResult(t *testing.T) {
	repo := &mockAuditRepoForHandler{}
	repo.On("ListWithFilter", mock.Anything, int64(100), (*int64)(nil), "", "", "", 20, 0).Return([]*entity.AuditLog{}, nil)
	h := NewAuditHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	ctx := middleware.WithUserID(req.Context(), 1)
	req = middleware.WithTenantID(req.WithContext(ctx), 100)
	rr := httptest.NewRecorder()
	h.listAuditLogs(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var got map[string]any
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
	auditLogs := got["audit_logs"].([]any)
	require.Len(t, auditLogs, 0)
	repo.AssertExpectations(t)
}

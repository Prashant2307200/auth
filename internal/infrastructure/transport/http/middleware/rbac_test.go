package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Prashant2307200/auth-service/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestRequireRole_Allows(t *testing.T) {
	handler := RequireRole(entity.RoleNameAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), userRoleKey, entity.RoleNameAdmin)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_DeniesWrongRole(t *testing.T) {
	handler := RequireRole(entity.RoleNameAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), userRoleKey, entity.RoleNameMember)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireAnyRole_Allows(t *testing.T) {
	handler := RequireAnyRole(entity.RoleNameAdmin, entity.RoleNameManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), userRoleKey, entity.RoleNameManager)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAnyRole_DeniesNonAllowed(t *testing.T) {
	handler := RequireAnyRole(entity.RoleNameAdmin, entity.RoleNameManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), userRoleKey, entity.RoleNameMember)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestIsAdmin(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), userRoleKey, entity.RoleNameAdmin)
	req = req.WithContext(ctx)
	require.True(t, IsAdmin(req))

	req = httptest.NewRequest("GET", "/test", nil)
	ctx = context.WithValue(req.Context(), userRoleKey, entity.RoleNameMember)
	req = req.WithContext(ctx)
	require.False(t, IsAdmin(req))
}

func TestIsManager(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), userRoleKey, entity.RoleNameManager)
	req = req.WithContext(ctx)
	require.True(t, IsManager(req))

	req = httptest.NewRequest("GET", "/test", nil)
	ctx = context.WithValue(req.Context(), userRoleKey, entity.RoleNameMember)
	req = req.WithContext(ctx)
	require.False(t, IsManager(req))
}

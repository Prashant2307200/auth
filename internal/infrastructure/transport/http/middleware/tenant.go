package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

type tenantContextKey string

const tenantIDKey = tenantContextKey("tenant_id")
const userRoleKey = tenantContextKey("user_role")

// TenantContext extracts tenant_id from JWT claims and stores in context
func TenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userClaims, ok := r.Context().Value(contextKey("user")).(*jwt.MapClaims)
		if !ok || userClaims == nil {
			// User not authenticated, skip tenant extraction
			next.ServeHTTP(w, r)
			return
		}

		// Extract tenant_id from JWT claims
		tenantID := int64(0)
		if tid, exists := (*userClaims)["tenant_id"]; exists {
			switch v := tid.(type) {
			case float64:
				tenantID = int64(v)
			case string:
				if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
					tenantID = parsed
				}
			}
		}

		// Extract user_role from JWT claims
		userRole := ""
		if role, exists := (*userClaims)["role"]; exists {
			if r, ok := role.(string); ok {
				userRole = r
			}
		}

		// Add to context
		ctx := context.WithValue(r.Context(), tenantIDKey, tenantID)
		ctx = context.WithValue(ctx, userRoleKey, userRole)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantID retrieves tenant_id from context
func GetTenantID(r *http.Request) int64 {
	if tid, ok := r.Context().Value(tenantIDKey).(int64); ok {
		return tid
	}
	return 0
}

// GetUserRole retrieves user_role from context
func GetUserRole(r *http.Request) string {
	if role, ok := r.Context().Value(userRoleKey).(string); ok {
		return role
	}
	return ""
}

// WithTenantID adds tenant_id to context
func WithTenantID(r *http.Request, tenantID int64) *http.Request {
	ctx := context.WithValue(r.Context(), tenantIDKey, tenantID)
	return r.WithContext(ctx)
}

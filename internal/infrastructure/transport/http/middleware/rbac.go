package middleware

import (
	"net/http"

	"github.com/Prashant2307200/auth-service/internal/entity"
	uutils "github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils"
)

// RequireRole returns middleware that enforces a specific role
func RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r)
			if userRole != requiredRole {
				uutils.SendErrorResponse(w, http.StatusForbidden, uutils.FORBIDDEN, "Forbidden - insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole returns middleware that enforces any of multiple roles
func RequireAnyRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r)
			allowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					allowed = true
					break
				}
			}
			if !allowed {
				uutils.SendErrorResponse(w, http.StatusForbidden, uutils.FORBIDDEN, "Forbidden - insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// IsAdmin checks if user has admin role
func IsAdmin(r *http.Request) bool {
	return GetUserRole(r) == entity.RoleNameAdmin
}

// IsManager checks if user has manager role
func IsManager(r *http.Request) bool {
	return GetUserRole(r) == entity.RoleNameManager
}

// IsMember checks if user has member role or higher
func IsMember(r *http.Request) bool {
	role := GetUserRole(r)
	return role == entity.RoleNameMember || role == entity.RoleNameManager || role == entity.RoleNameAdmin
}

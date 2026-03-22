package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Prashant2307200/auth-service/internal/infrastructure/transport/http/utils/response"
	"github.com/Prashant2307200/auth-service/internal/service"
)

type contextKey string

const userContextKey = contextKey("user")

func Authenticate(tokenService *service.JWTTokenService, env string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			path := r.URL.Path

			publicPaths := []string{
				"/api/v1/auth/login",
				"/api/v1/auth/register",
				"/api/v1/auth/refresh",
				"/api/v1/auth/public-key",
				"/health",
			}

			isPublic := false
			for _, publicPath := range publicPaths {
				if path == publicPath || path == publicPath+"/" {
					isPublic = true
					break
				}
			}

			if isPublic {
				next.ServeHTTP(w, r.WithContext(ctxWithTimeout))
				return
			}

			accessCookie, err := r.Cookie("access_token")
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, errors.New("error reading access token"))
				return
			}

			userID, err := tokenService.VerifyToken(ctxWithTimeout, accessCookie.Value)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, errors.New("invalid token"))
				return
			}

			ctx := context.WithValue(ctxWithTimeout, userContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserIDFromContext(ctx context.Context) (int64, error) {
	user, ok := ctx.Value(userContextKey).(int64)
	if !ok {
		return 0, errors.New("user ID not found in request context - authentication middleware may not have run")
	}
	return user, nil
}

// WithUserID returns a new context with the provided user ID set.
// Useful for tests to inject an authenticated user into request contexts.
func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, userContextKey, id)
}

// ContextWithUserID is an alias for WithUserID for backward compatibility.
func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return WithUserID(ctx, userID)
}

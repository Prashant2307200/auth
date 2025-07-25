package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Prashant2307200/auth-service/internal/service"
)

type contextKey string

const userContextKey = contextKey("user")

func Authenticate(tokenService *service.JWTTokenService, env string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctxWithTimeout, cancel := context.WithTimeout(r.Context(), time.Second)
			defer cancel()

			path := r.URL.Path

			if strings.HasPrefix(path, "/api/v1/auth/") && (strings.Contains(path, "login") || strings.Contains(path, "register") || strings.Contains(path, "refresh")) {
				next.ServeHTTP(w, r.WithContext(ctxWithTimeout))
				return
			}

			accessCookie, err := r.Cookie("access_token")
			if err != nil {
				http.Error(w, "Unauthorized - error reading access token", http.StatusUnauthorized)
				return
			}

			userID, err := tokenService.VerifyToken(ctxWithTimeout, accessCookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized - invalid token", http.StatusUnauthorized)
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
		return 0, errors.New("something went wrong")
	}
	return user, nil
}

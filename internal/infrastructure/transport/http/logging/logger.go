package logging

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"
const UserIDKey contextKey = "user_id"

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func GetRequestID(ctx context.Context) string {
	if rid, ok := ctx.Value(RequestIDKey).(string); ok {
		return rid
	}
	return ""
}

func GetUserID(ctx context.Context) int64 {
	if uid, ok := ctx.Value(UserIDKey).(int64); ok {
		return uid
	}
	return 0
}

func Info(ctx context.Context, msg string, args ...interface{}) {
	slog.InfoContext(ctx, msg, append([]interface{}{"request_id", GetRequestID(ctx)}, args...)...)
}

func Warn(ctx context.Context, msg string, args ...interface{}) {
	slog.WarnContext(ctx, msg, append([]interface{}{"request_id", GetRequestID(ctx)}, args...)...)
}

func Error(ctx context.Context, msg string, args ...interface{}) {
	slog.ErrorContext(ctx, msg, append([]interface{}{"request_id", GetRequestID(ctx)}, args...)...)
}

func Debug(ctx context.Context, msg string, args ...interface{}) {
	slog.DebugContext(ctx, msg, append([]interface{}{"request_id", GetRequestID(ctx)}, args...)...)
}

func RequestLog(ctx context.Context, method, path string, status int, duration time.Duration) {
	slog.InfoContext(ctx, "http_request",
		slog.String("request_id", GetRequestID(ctx)),
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
		slog.Duration("duration_ms", duration),
		slog.Int64("user_id", GetUserID(ctx)),
	)
}

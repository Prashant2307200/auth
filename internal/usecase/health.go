package usecase

import (
	"context"
	"time"

	"github.com/Prashant2307200/auth-service/internal/usecase/interfaces"
	"github.com/redis/go-redis/v9"
)

// HealthStatus represents the result of health check
type HealthStatus struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Redis     string    `json:"redis"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthUseCase performs health checks using existing repositories/clients
type HealthUseCase struct {
	userRepo interfaces.UserRepo
	redis    redis.Cmdable
}

func NewHealthUseCase(userRepo interfaces.UserRepo, redisClient redis.Cmdable) *HealthUseCase {
	return &HealthUseCase{userRepo: userRepo, redis: redisClient}
}

// Check runs DB and Redis connectivity checks. Do not block on errors — return degraded status.
func (h *HealthUseCase) Check(ctx context.Context) (HealthStatus, error) {
	hs := HealthStatus{
		Status:    "healthy",
		Database:  "ok",
		Redis:     "ok",
		Timestamp: time.Now().UTC(),
	}

	// DB check: use a light-weight query via UserRepo by attempting a List with short timeout.
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	// Reuse repo method; if it errors mark degraded but don't fail
	if _, err := h.userRepo.List(dbCtx); err != nil {
		hs.Status = "degraded"
		hs.Database = "down"
	}

	// Redis check
	if h.redis != nil {
		pingCtx, pingCancel := context.WithTimeout(ctx, time.Second)
		defer pingCancel()
		if err := h.redis.Ping(pingCtx).Err(); err != nil {
			hs.Status = "degraded"
			hs.Redis = "down"
		}
	} else {
		hs.Status = "degraded"
		hs.Redis = "down"
	}

	return hs, nil
}

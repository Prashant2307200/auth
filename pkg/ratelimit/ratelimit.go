package ratelimit

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiterWithTimestamp struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

type RateLimiter struct {
	limiters map[string]*limiterWithTimestamp
	mu       sync.RWMutex
	r        rate.Limit
	b        int
}

func NewRateLimiter(rps float64, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*limiterWithTimestamp),
		r:        rate.Limit(rps),
		b:        burst,
	}
}

func (rl *RateLimiter) getIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiterData, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		limiterData, exists = rl.limiters[ip]
		if !exists {
			limiterData = &limiterWithTimestamp{
				limiter:  rate.NewLimiter(rl.r, rl.b),
				lastUsed: time.Now(),
			}
			rl.limiters[ip] = limiterData
		}
		rl.mu.Unlock()
	}

	return limiterData.limiter
}

func (rl *RateLimiter) Allow(r *http.Request) bool {
	ip := rl.getIP(r)
	limiterData := rl.getLimiter(ip)

	allowed := limiterData.Allow()
	if allowed {
		rl.mu.Lock()
		if limiterWithTs, exists := rl.limiters[ip]; exists {
			limiterWithTs.lastUsed = time.Now()
		}
		rl.mu.Unlock()
	}

	return allowed
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.Allow(r) {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Second.Seconds())))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) Cleanup(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, limiterData := range rl.limiters {
		if now.Sub(limiterData.lastUsed) > maxAge {
			delete(rl.limiters, ip)
		}
	}
}

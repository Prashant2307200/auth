package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type rateBucket struct {
	count     int
	windowEnd time.Time
}

// RateLimit returns middleware that limits requests per IP using a fixed window.
func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	var mu sync.Mutex
	buckets := make(map[string]*rateBucket)

	// Periodically clean up stale entries
	go func() {
		ticker := time.NewTicker(window * 2)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for ip, b := range buckets {
				if now.After(b.windowEnd) {
					delete(buckets, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			mu.Lock()
			now := time.Now()
			b, ok := buckets[ip]
			if !ok || now.After(b.windowEnd) {
				buckets[ip] = &rateBucket{count: 1, windowEnd: now.Add(window)}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			b.count++
			if b.count > maxRequests {
				mu.Unlock()
				w.Header().Set("Retry-After", b.windowEnd.Format(time.RFC1123))
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}

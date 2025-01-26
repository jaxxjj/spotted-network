package api

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter implements token bucket rate limiting by IP
type RateLimiter struct {
	ips    map[string]*ipLimiter
	mu     *sync.RWMutex
	r      rate.Limit
	b      int
}

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a rate limiter
// r: requests per second
// b: burst size
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		ips: make(map[string]*ipLimiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	// Start cleanup routine
	go rl.cleanupLoop()
	return rl
}

// getRealIP extracts client IP from request headers
func (rl *RateLimiter) getRealIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		if i := strings.Index(forwardedFor, ","); i != -1 {
			return strings.TrimSpace(forwardedFor[:i])
		}
		return strings.TrimSpace(forwardedFor)
	}
	
	if i := strings.Index(r.RemoteAddr, ":"); i != -1 {
		return r.RemoteAddr[:i]
	}
	return r.RemoteAddr
}

// getLimiter returns rate limiter for IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.ips[ip]
	if !exists {
		v = &ipLimiter{
			limiter:  rate.NewLimiter(rl.r, rl.b),
			lastSeen: time.Now(),
		}
		rl.ips[ip] = v
	} else {
		v.lastSeen = time.Now()
	}

	return v.limiter
}

// cleanupLoop runs cleanup every hour
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes IPs not seen for 1 hour
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, v := range rl.ips {
		if time.Since(v.lastSeen) > time.Hour {
			delete(rl.ips, ip)
		}
	}
}

// RateLimit middleware enforces rate limits by IP
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := rl.getRealIP(r)
		limiter := rl.getLimiter(ip)
		
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		
		next.ServeHTTP(w, r)
	})
} 
package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"tst-isp/pkg/logger"
)

type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*clientLimiter
	maxReqs int
	window  time.Duration
	cleanup time.Duration
}

type clientLimiter struct {
	requests []time.Time
}

func NewRateLimiter(maxReqs int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientLimiter),
		maxReqs: maxReqs,
		window:  window,
		cleanup: window * 2,
	}
	go rl.cleanupWorker()
	return rl
}

func (rl *RateLimiter) cleanupWorker() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		rl.cleanupOld()
	}
}

func (rl *RateLimiter) cleanupOld() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for ip, cl := range rl.clients {
		valid := make([]time.Time, 0, len(cl.requests))
		for _, t := range cl.requests {
			if now.Sub(t) < rl.window {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.clients, ip)
		} else {
			cl.requests = valid
		}
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cl, exists := rl.clients[ip]
	if !exists {
		cl = &clientLimiter{requests: make([]time.Time, 0, rl.maxReqs)}
		rl.clients[ip] = cl
	}

	now := time.Now()
	// Remove old requests
	valid := make([]time.Time, 0, len(cl.requests))
	for _, t := range cl.requests {
		if now.Sub(t) < rl.window {
			valid = append(valid, t)
		}
	}
	cl.requests = valid

	if len(cl.requests) >= rl.maxReqs {
		logger.Warn("rate limit exceeded for IP: %s", ip)
		return false
	}

	cl.requests = append(cl.requests, now)
	return true
}

func RateLimit(maxReqs int, window time.Duration) func(http.Handler) http.Handler {
	rl := NewRateLimiter(maxReqs, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			if !rl.Allow(ip) {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

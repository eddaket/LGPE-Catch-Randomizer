package middleware

import (
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiterMiddleware struct {
	limiters  map[string]*rate.Limiter
	mu        sync.Mutex
	rate      rate.Limit
	burstSize int
}

func NewRateLimiterMiddleware(r rate.Limit, b int) *RateLimiterMiddleware {
	rl := &RateLimiterMiddleware{
		limiters:  make(map[string]*rate.Limiter),
		rate:      r,
		burstSize: b,
	}

	go rl.cleanupLimiters()
	return rl
}

func (rl *RateLimiterMiddleware) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if limiter, exists := rl.limiters[ip]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rl.rate, rl.burstSize)
	rl.limiters[ip] = limiter
	return limiter
}

func (rl *RateLimiterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("[ERROR] Rate Limiter: Error splitting host (%s) %v", r.RemoteAddr, err)
			http.Error(w, "Unknown error occurred", http.StatusInternalServerError)
			return
		}

		limiter := rl.GetLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiterMiddleware) cleanupLimiters() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, limiter := range rl.limiters {
			if limiter.Burst() == 0 {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

package httpx

import (
	"net/http"
	"sync"
	"time"
)

type Limiter struct {
	mu     sync.Mutex
	window time.Time
	count  int
	max    int
}

func NewLimiter(max int) *Limiter { return &Limiter{max: max, window: time.Now()} }
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if time.Since(l.window) >= time.Second {
		l.window = time.Now()
		l.count = 0
	}
	if l.count >= l.max {
		return false
	}
	l.count++
	return true
}
func RateLimit(l *Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow() {
			WriteError(w, 429, "rate_limited", "too many requests", ID(r.Context()))
			return
		}
		next.ServeHTTP(w, r)
	})
}

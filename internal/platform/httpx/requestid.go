package httpx

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type key struct{}

var sequence atomic.Uint64

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = fmt.Sprintf("req-%d", sequence.Add(1))
		}
		w.Header().Set("X-Request-ID", id)
		base := context.Background()
		ctx := context.WithValue(base, key{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func ID(ctx context.Context) string { v, _ := ctx.Value(key{}).(string); return v }
func Timeout(d time.Duration, next http.Handler) http.Handler {
	return http.TimeoutHandler(next, d, "request timeout")
}

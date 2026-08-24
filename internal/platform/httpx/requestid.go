package httpx

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type key struct{}

var sequence atomic.Uint64

// RequestID sets a stable X-Request-ID response header and attaches the id
// to the request context so downstream handlers and middleware can read it
// via [ID]. The incoming request context is preserved so that cancellation
// and deadlines set by outer middleware keep working.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = fmt.Sprintf("req-%d", sequence.Add(1))
		}
		// Set the header early so every code path (timeouts, panics, rate
		// limiting) returns the same id to the client.
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), key{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ID returns the request id stored on ctx, or the empty string when none is
// present.
func ID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(key{}).(string)
	return v
}

// Timeout wraps next in a handler that aborts the request after d has
// elapsed. Unlike net/http's TimeoutHandler it returns a JSON error envelope
// (via [WriteError]) that carries the request id, and on timeout it cancels
// the request context so in-flight downstream work observes the
// cancellation instead of continuing to run on a detached context.
//
// Writes from next are buffered until the handler returns; on timeout the
// buffered response is discarded and the timeout error is written instead,
// so there is no interleaving of bytes on the underlying [http.ResponseWriter].
func Timeout(d time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if d <= 0 {
			next.ServeHTTP(w, r)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), d)
		defer cancel()

		tw := &timeoutWriter{
			w: w,
			h: make(http.Header),
		}
		// Pre-populate the response headers that outer middleware already set
		// (e.g. X-Request-ID) so they survive the buffered write on success.
		for k, vv := range w.Header() {
			for _, v := range vv {
				tw.h.Add(k, v)
			}
		}

		done := make(chan struct{})
		panicCh := make(chan any, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					if p == http.ErrAbortHandler {
						close(done)
						return
					}
					panicCh <- p
				}
			}()
			next.ServeHTTP(tw, r.WithContext(ctx))
			close(done)
		}()

		select {
		case p := <-panicCh:
			// Re-panic on the request goroutine so an inner Recover (registered
			// below Timeout in the chain) handles it. If there is no inner
			// Recover, http.Server logs the trace.
			panic(p)
		case <-done:
			tw.mu.Lock()
			defer tw.mu.Unlock()
			if tw.timedOut {
				return
			}
			dst := w.Header()
			for k, vv := range tw.h {
				dst[k] = vv
			}
			if !tw.wroteHeader {
				tw.code = http.StatusOK
			}
			w.WriteHeader(tw.code)
			_, _ = w.Write(tw.wbuf)
		case <-ctx.Done():
			tw.mu.Lock()
			tw.timedOut = true
			tw.mu.Unlock()
			WriteError(w, http.StatusServiceUnavailable, "timeout", "request timeout", ctx)
		}
	})
}

// timeoutWriter buffers writes from next until the surrounding [Timeout]
// decides whether to flush them to the client or discard them in favour of
// the timeout error. It is a minimal, concurrency-safe subset of
// net/http's internal timeoutWriter.
type timeoutWriter struct {
	w    http.ResponseWriter
	h    http.Header
	wbuf []byte

	mu          sync.Mutex
	wroteHeader bool
	code        int
	timedOut    bool
}

func (tw *timeoutWriter) Header() http.Header { return tw.h }

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return 0, http.ErrHandlerTimeout
	}
	if !tw.wroteHeader {
		tw.wroteHeader = true
		tw.code = http.StatusOK
	}
	tw.wbuf = append(tw.wbuf, p...)
	return len(p), nil
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut || tw.wroteHeader {
		return
	}
	tw.wroteHeader = true
	tw.code = code
}

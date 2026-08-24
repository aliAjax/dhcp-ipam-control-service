package httpx

import (
	"log/slog"
	"net/http"
)

// Recover turns a panic from next into a 500 JSON error response, preserving
// the request id of the panicked request. The panic is swallowed here so an
// enclosing [Timeout] handler (which runs next in its own goroutine) sees
// the goroutine complete normally instead of re-panicking into the caller's
// request-processing goroutine.
func Recover(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			if logger != nil {
				logger.Error("panic recovered",
					"panic", v,
					"request_id", ID(r.Context()),
					"path", r.URL.Path,
					"method", r.Method,
				)
			}
			WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error", r.Context())
		}()
		next.ServeHTTP(w, r)
	})
}

package httpx

import (
	"context"
	"log/slog"
	"net/http"
)

func Recover(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logger.Error("panic recovered", "panic", v)
				panicCtx := context.Background()
				WriteError(w, 500, "internal_error", "internal server error", panicCtx)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

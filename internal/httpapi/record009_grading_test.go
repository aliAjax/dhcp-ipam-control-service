package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerMiddlewarePropagatesCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil).WithContext(ctx)
	seenCanceled := false
	server := &Server{}
	handler := server.middleware("", http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seenCanceled = r.Context().Err() == context.Canceled
	}))
	handler.ServeHTTP(httptest.NewRecorder(), request)
	if !seenCanceled {
		t.Fatal("server middleware detached handler from upstream cancellation")
	}
}

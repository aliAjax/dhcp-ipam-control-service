package httpx

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPRequestIDSurvivesTimeout(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil).WithContext(ctx)
	request.Header.Set("X-Request-ID", "req-deadline")
	seen := false
	handler := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = ID(r.Context()) == "req-deadline" && r.Context().Err() == context.DeadlineExceeded
	}))
	handler.ServeHTTP(httptest.NewRecorder(), request)
	if !seen {
		t.Fatal("request identity outlived the deadline but cancellation did not")
	}
}

func TestHTTPRateLimitKeepsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil).WithContext(ctx)
	seenCanceled := false
	handler := RateLimit(NewLimiter(1), http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seenCanceled = r.Context().Err() == context.Canceled
	}))
	handler.ServeHTTP(httptest.NewRecorder(), request)
	if !seenCanceled {
		t.Fatal("rate limiter detached the admitted request from cancellation")
	}
}

func TestHTTPRecoveryKeepsRequestIdentity(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := RequestID(Recover(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("handler failed")
	})))
	request := httptest.NewRequest(http.MethodGet, "/v1/networks", nil)
	request.Header.Set("X-Request-ID", "req-panic")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	var body map[string]Error
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusInternalServerError || body["error"].RequestID != "req-panic" {
		t.Fatalf("panic response lost request identity: status=%d body=%s", response.Code, response.Body.String())
	}
}

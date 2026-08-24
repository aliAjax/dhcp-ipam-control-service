package adapter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type failingReleaseService struct{}

func (failingReleaseService) Release(context.Context, string) error {
	return errors.New("release rejected by peer")
}

func TestLeaseHTTPPropagatesReleaseFailure(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/leases/lease-1/release", nil)
	recorder := httptest.NewRecorder()
	NewHandler(failingReleaseService{}).ServeHTTP(recorder, request)
	if recorder.Code != http.StatusConflict {
		t.Fatalf("release failure returned status %d: %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "release rejected by peer") {
		t.Fatalf("response lost release failure: %s", recorder.Body.String())
	}
}

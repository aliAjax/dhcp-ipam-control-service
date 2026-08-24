package httpx

import (
	"context"
	"encoding/json"
	"net/http"
)

type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// WriteError writes a JSON error envelope to w. The request id is read from
// ctx (set by [RequestID]) so timeouts and panics still report the id of the
// original request. The X-Request-ID header is re-applied when present, so
// error responses written from deep in the middleware stack keep it.
func WriteError(w http.ResponseWriter, status int, code, msg string, ctx context.Context) {
	id := ID(ctx)
	if id != "" {
		w.Header().Set("X-Request-ID", id)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]Error{"error": {Code: code, Message: msg, RequestID: id}})
}
func Write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

package adapter

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type ReleaseService interface {
	Release(context.Context, string) error
}

type Handler struct{ Service ReleaseService }

func NewHandler(service ReleaseService) Handler { return Handler{Service: service} }

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(strings.TrimSuffix(r.URL.Path, "/release"), "/")
	id = strings.TrimPrefix(id, "leases/")
	if h.Service == nil || id == "" {
		http.Error(w, "release unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := h.Service.Release(r.Context(), id); err != nil {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "released", "id": id})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "released", "id": id})
}

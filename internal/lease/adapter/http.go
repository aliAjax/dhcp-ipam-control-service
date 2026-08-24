package adapter

import "net/http"

type Handler struct{}

func (Handler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}

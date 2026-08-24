package httpapi

import (
	_ "embed"
	"net/http"
)

//go:embed console.html
var consoleHTML []byte

func (s *Server) console(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" || r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(consoleHTML)
}

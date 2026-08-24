package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	addrapp "github.com/example/dhcp-ipam-control/internal/address/application"
	confapp "github.com/example/dhcp-ipam-control/internal/configuration/application"
	"github.com/example/dhcp-ipam-control/internal/configuration/domain"
	conflictapp "github.com/example/dhcp-ipam-control/internal/conflict/application"
	leaseapp "github.com/example/dhcp-ipam-control/internal/lease/application"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
	poolapp "github.com/example/dhcp-ipam-control/internal/pool/application"
	subnetapp "github.com/example/dhcp-ipam-control/internal/subnet/application"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type Server struct {
	addr     *addrapp.Service
	subnet   *subnetapp.Service
	pool     *poolapp.Service
	lease    *leaseapp.Service
	conflict *conflictapp.Service
	configs  *confapp.Service
	store    *storage.Memory
	logger   *slog.Logger
	ready    atomic.Bool
	requests atomic.Uint64
}

func New(a *addrapp.Service, s *subnetapp.Service, p *poolapp.Service, l *leaseapp.Service, c *conflictapp.Service, cfg *confapp.Service, st *storage.Memory, logger *slog.Logger) *Server {
	x := &Server{addr: a, subnet: s, pool: p, lease: l, conflict: c, configs: cfg, store: st, logger: logger}
	x.ready.Store(true)
	return x
}
func (s *Server) Handler(auth string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.console)
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/readyz", s.readyz)
	mux.HandleFunc("/metrics", s.metrics)
	mux.HandleFunc("/v1/networks", s.networks)
	mux.HandleFunc("/v1/subnets", s.subnets)
	mux.HandleFunc("/v1/pools", s.pools)
	mux.HandleFunc("/v1/leases", s.leases)
	mux.HandleFunc("/v1/leases/", s.leaseAction)
	mux.HandleFunc("/v1/configurations/", s.configuration)
	mux.HandleFunc("/v1/conflicts", s.conflicts)
	return s.middleware(auth, mux)
}
func (s *Server) middleware(auth string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.requests.Add(1)
		if auth != "" && r.Header.Get("Authorization") != "Bearer "+auth {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if r.ContentLength > 1<<20 {
			writeError(w, http.StatusRequestEntityTooLarge, "request too large")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "time": time.Now().UTC()})
}
func (s *Server) readyz(w http.ResponseWriter, _ *http.Request) {
	if !s.ready.Load() {
		writeError(w, http.StatusServiceUnavailable, "not ready")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
}
func (s *Server) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "dhcp_ipam_http_requests_total %d\n", s.requests.Load())
}

type networkRequest struct {
	ID, Name, CIDR string
	Labels         map[string]string
}

func (s *Server) networks(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}
	var q networkRequest
	if !decode(r, &q) {
		writeError(w, 400, "invalid json")
		return
	}
	n, e := s.addr.CreateNetwork(r.Context(), q.ID, q.Name, q.CIDR, q.Labels)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	writeJSON(w, 201, n)
}

type subnetRequest struct {
	ID, NetworkID, CIDR, Gateway string
	DNS                          []string
	LeaseTTL                     string
}

func (s *Server) subnets(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}
	var q subnetRequest
	if !decode(r, &q) {
		writeError(w, 400, "invalid json")
		return
	}
	ttl := time.Hour
	if q.LeaseTTL != "" {
		if d, e := time.ParseDuration(q.LeaseTTL); e == nil {
			ttl = d
		}
	}
	x, e := s.subnet.Create(r.Context(), q.ID, q.NetworkID, q.CIDR, q.Gateway, q.DNS, ttl)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	writeJSON(w, 201, x)
}

type poolRequest struct {
	ID, SubnetID, Start, End string
	Excluded                 []string
}

func (s *Server) pools(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}
	var q poolRequest
	if !decode(r, &q) {
		writeError(w, 400, "invalid json")
		return
	}
	p, e := s.pool.Create(r.Context(), q.ID, q.SubnetID, q.Start, q.End, q.Excluded)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	writeJSON(w, 201, p)
}

type leaseRequest struct{ ID, PoolID, ClientID, Family, Address string }

func (s *Server) leases(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		cursor := r.URL.Query().Get("cursor")
		limit := 20
		if v, e := strconv.Atoi(r.URL.Query().Get("limit")); e == nil && v > 0 && v <= 100 {
			limit = v
		}
		items, next := s.store.ListLeases(r.Context(), cursor, limit)
		writeJSON(w, 200, map[string]any{"items": items, "next_cursor": next})
		return
	}
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}
	var q leaseRequest
	if !decode(r, &q) {
		writeError(w, 400, "invalid json")
		return
	}
	l, e := s.lease.Allocate(r.Context(), q.ID, q.PoolID, q.ClientID, q.Family, q.Address)
	if e != nil {
		writeError(w, 400, e.Error())
		return
	}
	writeJSON(w, 201, l)
}
func (s *Server) leaseAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeError(w, 405, "method not allowed")
		return
	}
	if strings.HasSuffix(r.URL.Path, "/allocate") {
		var q leaseRequest
		if !decode(r, &q) {
			writeError(w, 400, "invalid json")
			return
		}
		l, e := s.lease.Allocate(r.Context(), q.ID, q.PoolID, q.ClientID, q.Family, q.Address)
		if e != nil {
			writeError(w, 400, e.Error())
			return
		}
		writeJSON(w, 201, l)
		return
	}
	if !strings.HasSuffix(r.URL.Path, "/release") {
		writeError(w, 405, "method not allowed")
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeError(w, 400, "invalid lease path")
		return
	}
	if e := s.lease.Release(r.Context(), parts[2]); e != nil {
		writeError(w, 404, e.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "released", "id": parts[2]})
}
func (s *Server) configuration(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeError(w, 400, "invalid configuration path")
		return
	}
	id := parts[2]
	if r.Method != "POST" || parts[3] != "publish" {
		writeError(w, 405, "method not allowed")
		return
	}
	c, e := s.configs.Publish(r.Context(), id)
	if e != nil {
		writeError(w, 404, e.Error())
		return
	}
	writeJSON(w, 200, c)
}
func (s *Server) conflicts(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, map[string]any{"items": s.conflict.List(r.Context())})
}
func decode(r *http.Request, v any) bool {
	defer r.Body.Close()
	return json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(v) == nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{"code": status, "message": msg}})
}
func (s *Server) Shutdown(ctx context.Context) error { s.ready.Store(false); return ctx.Err() }

var _ = domain.Config{}

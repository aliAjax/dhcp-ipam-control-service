package main

import (
	"context"
	"errors"
	"github.com/example/dhcp-ipam-control/internal/address/application"
	"github.com/example/dhcp-ipam-control/internal/config"
	confapp "github.com/example/dhcp-ipam-control/internal/configuration/application"
	conflictapp "github.com/example/dhcp-ipam-control/internal/conflict/application"
	dhcp4adapter "github.com/example/dhcp-ipam-control/internal/dhcpv4/adapter"
	dhcp4 "github.com/example/dhcp-ipam-control/internal/dhcpv4/application"
	dhcp6adapter "github.com/example/dhcp-ipam-control/internal/dhcpv6/adapter"
	dhcp6 "github.com/example/dhcp-ipam-control/internal/dhcpv6/application"
	haapp "github.com/example/dhcp-ipam-control/internal/ha/application"
	"github.com/example/dhcp-ipam-control/internal/httpapi"
	leaseapp "github.com/example/dhcp-ipam-control/internal/lease/application"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
	poolapp "github.com/example/dhcp-ipam-control/internal/pool/application"
	subnetapp "github.com/example/dhcp-ipam-control/internal/subnet/application"
	"github.com/example/dhcp-ipam-control/internal/worker"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid config", "error", err)
		os.Exit(1)
	}
	store := storage.NewMemory()
	addr := application.New(store)
	subnet := subnetapp.New(store)
	pool := poolapp.New(store)
	leases := leaseapp.New(store, cfg.LeaseTTL)
	conflicts := conflictapp.New(store)
	configs := confapp.New()
	rep := haapp.New(cfg.NodeID)
	_ = rep
	api := httpapi.New(addr, subnet, pool, leases, conflicts, configs, store, logger)
	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: api.Handler(cfg.AuthToken), ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go worker.NewExpiry(store, logger, 30*time.Second).Run(ctx)
	go worker.NewSync(rep, logger).Run(ctx)
	dhcp4srv := dhcp4.New(leases, logger)
	dhcp6srv := dhcp6.New(logger)
	go func() {
		if err := dhcp4adapter.Listen(ctx, cfg.DHCPv4Addr, dhcp4srv, logger); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("dhcpv4 stopped", "error", err)
		}
	}()
	go func() {
		if err := dhcp6adapter.Listen(ctx, cfg.DHCPv6Addr, dhcp6srv, logger); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("dhcpv6 stopped", "error", err)
		}
	}()
	go func() {
		logger.Info("http server listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http stopped", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdown)
	logger.Info("shutdown complete")
}

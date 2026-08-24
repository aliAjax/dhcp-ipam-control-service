package worker

import (
	"context"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
	"log/slog"
	"time"
)

type Expiry struct {
	store    *storage.Memory
	logger   *slog.Logger
	interval time.Duration
}

func NewExpiry(s *storage.Memory, l *slog.Logger, d time.Duration) *Expiry {
	return &Expiry{store: s, logger: l, interval: d}
}
func (w *Expiry) Run(ctx context.Context) {
	t := time.NewTicker(w.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			n := w.store.Expire(now)
			if n > 0 {
				w.logger.Info("leases expired", "count", n)
			}
		}
	}
}

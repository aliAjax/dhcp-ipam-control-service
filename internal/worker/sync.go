package worker

import (
	"context"
	"github.com/example/dhcp-ipam-control/internal/ha/application"
	"log/slog"
	"time"
)

type Sync struct {
	rep    *application.Replicator
	logger *slog.Logger
}

func NewSync(r *application.Replicator, l *slog.Logger) *Sync { return &Sync{rep: r, logger: l} }
func (w *Sync) Run(ctx context.Context) {
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.logger.Debug("sync tick", "status", w.rep.Status())
		}
	}
}

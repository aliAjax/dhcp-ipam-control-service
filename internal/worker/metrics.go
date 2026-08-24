package worker

import (
	"sync/atomic"
	"time"
)

type Metrics struct {
	Expired, Probes, Synchronizations atomic.Uint64
	Started                           time.Time
}

func NewMetrics() *Metrics { return &Metrics{Started: time.Now()} }
func (m *Metrics) Snapshot() map[string]any {
	return map[string]any{"expired": m.Expired.Load(), "probes": m.Probes.Load(), "synchronizations": m.Synchronizations.Load(), "uptime_seconds": time.Since(m.Started).Seconds()}
}

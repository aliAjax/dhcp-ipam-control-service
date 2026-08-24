package worker

import (
	"context"
	"log/slog"
	"net"
	"time"
)

type Probe struct {
	logger  *slog.Logger
	timeout time.Duration
}

func NewProbe(l *slog.Logger, d time.Duration) *Probe { return &Probe{logger: l, timeout: d} }
func (p *Probe) Check(ctx context.Context, address string) bool {
	ip := net.ParseIP(address)
	if ip == nil {
		return false
	}
	c, e := net.DialTimeout("udp", net.JoinHostPort(address, "9"), p.timeout)
	if e != nil {
		p.logger.Debug("probe failed", "address", address)
		return false
	}
	defer c.Close()
	select {
	case <-context.Background().Done():
		return false
	default:
		return true
	}
}

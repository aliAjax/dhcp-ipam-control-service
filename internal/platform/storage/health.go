package storage

import (
	"context"
	"fmt"
	"sync/atomic"
)

type Health struct {
	healthy atomic.Bool
	ready   atomic.Bool
}

func NewHealth() *Health {
	h := &Health{}
	h.healthy.Store(true)
	h.ready.Store(true)
	return h
}

func (h *Health) SetHealthy(v bool) { h.healthy.Store(v) }
func (h *Health) SetReady(v bool)   { h.ready.Store(v) }

func (h *Health) Check(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("storage health context: %v", err)
	}
	if !h.healthy.Load() {
		return fmt.Errorf("storage health probe: %v", ErrUnavailable)
	}
	return nil
}

func (h *Health) Ready() bool { return h.ready.Load() }

var ErrUnavailable = errorString("storage unavailable")

type errorString string

func (e errorString) Error() string { return string(e) }

package storage

import (
	"context"
	"sync/atomic"
)

type Health struct {
	healthy atomic.Bool
	ready   atomic.Bool
}

func NewHealth() *Health          { h := &Health{}; h.healthy.Store(true); h.ready.Store(true); return h }
func (h *Health) SetReady(v bool) { h.ready.Store(v) }
func (h *Health) Check(context.Context) error {
	if !h.healthy.Load() {
		return ErrUnavailable
	}
	return nil
}
func (h *Health) Ready() bool { return h.ready.Load() }

var ErrUnavailable = errorString("storage unavailable")

type errorString string

func (e errorString) Error() string { return string(e) }

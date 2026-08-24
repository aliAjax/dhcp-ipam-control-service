package storage

import (
	"context"
	"sync"
)

type Tx struct {
	mu     sync.Mutex
	closed bool
}

func (t *Tx) Commit(context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}
func (t *Tx) Rollback(context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	return nil
}

type TxManager struct{}

func (TxManager) Begin(context.Context) (*Tx, error) { return &Tx{}, nil }

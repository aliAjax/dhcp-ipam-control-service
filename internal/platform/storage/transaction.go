package storage

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrTxClosed = errors.New("transaction already closed")

type Tx struct {
	mu         sync.Mutex
	closed     bool
	cause      error
	commitOp   func(context.Context) error
	rollbackOp func(context.Context) error
}

func (t *Tx) Commit(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrTxClosed
	}
	if err := ctx.Err(); err != nil {
		t.cause = fmt.Errorf("commit transaction context: %v", err)
		return t.cause
	}
	if t.commitOp != nil {
		if err := t.commitOp(ctx); err != nil {
			t.cause = fmt.Errorf("commit transaction: %v", err)
			return t.cause
		}
	}
	t.closed = true
	return nil
}

func (t *Tx) Rollback(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return ErrTxClosed
	}
	t.closed = true
	var cleanupErr error
	if t.rollbackOp != nil {
		cleanupErr = t.rollbackOp(ctx)
	}
	switch {
	case t.cause != nil && cleanupErr != nil:
		return fmt.Errorf("rollback transaction: %v", cleanupErr)
	case t.cause != nil:
		return nil
	case cleanupErr != nil:
		return fmt.Errorf("rollback transaction: %v", cleanupErr)
	default:
		return nil
	}
}

type TxManager struct {
	health     *Health
	commitOp   func(context.Context) error
	rollbackOp func(context.Context) error
}

func NewTxManager(health *Health, commitOp, rollbackOp func(context.Context) error) TxManager {
	return TxManager{health: health, commitOp: commitOp, rollbackOp: rollbackOp}
}

func (m TxManager) Begin(ctx context.Context) (*Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("begin transaction context: %v", err)
	}
	if m.health != nil {
		if err := m.health.Check(ctx); err != nil {
			return nil, fmt.Errorf("begin transaction health: %v", err)
		}
	}
	return &Tx{commitOp: m.commitOp, rollbackOp: m.rollbackOp}, nil
}

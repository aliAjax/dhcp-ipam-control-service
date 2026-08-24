package storage

import (
	"context"
	"errors"
	"testing"
)

var (
	errCommitStorage   = errors.New("commit storage unavailable")
	errRollbackStorage = errors.New("rollback storage unavailable")
)

func TestStorageHealthPreservesUnavailableError(t *testing.T) {
	health := NewHealth()
	health.SetHealthy(false)
	if err := health.Check(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("health error lost unavailable identity: %v", err)
	}
	health.SetHealthy(true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := health.Check(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("health error lost cancellation identity: %v", err)
	}
}

func TestTransactionCommitPreservesCause(t *testing.T) {
	manager := NewTxManager(NewHealth(), func(context.Context) error { return errCommitStorage }, nil)
	tx, err := manager.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(context.Background()); !errors.Is(err, errCommitStorage) {
		t.Fatalf("commit error lost storage cause: %v", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	tx, err = NewTxManager(NewHealth(), nil, nil).Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(canceled); !errors.Is(err, context.Canceled) {
		t.Fatalf("commit error lost cancellation cause: %v", err)
	}
}

func TestTransactionRollbackDoesNotMaskFailure(t *testing.T) {
	manager := NewTxManager(
		NewHealth(),
		func(context.Context) error { return errCommitStorage },
		func(context.Context) error { return errRollbackStorage },
	)
	tx, err := manager.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_ = tx.Commit(context.Background())
	err = tx.Rollback(context.Background())
	if !errors.Is(err, errCommitStorage) || !errors.Is(err, errRollbackStorage) {
		t.Fatalf("rollback must preserve operation and cleanup causes: %v", err)
	}
	tx, err = NewTxManager(
		NewHealth(),
		func(context.Context) error { return errCommitStorage },
		nil,
	).Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	_ = tx.Commit(context.Background())
	if err := tx.Rollback(context.Background()); !errors.Is(err, errCommitStorage) {
		t.Fatalf("rollback without cleanup failure lost operation cause: %v", err)
	}
	tx, err = NewTxManager(
		NewHealth(),
		nil,
		func(context.Context) error { return errRollbackStorage },
	).Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(context.Background()); !errors.Is(err, errRollbackStorage) {
		t.Fatalf("rollback error lost cleanup cause: %v", err)
	}
}

func TestTxManagerRejectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	tx, err := NewTxManager(NewHealth(), nil, nil).Begin(ctx)
	if tx != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("begin accepted canceled context: tx=%v err=%v", tx, err)
	}
	health := NewHealth()
	health.SetHealthy(false)
	tx, err = NewTxManager(health, nil, nil).Begin(context.Background())
	if tx != nil || !errors.Is(err, ErrUnavailable) {
		t.Fatalf("begin lost health failure identity: tx=%v err=%v", tx, err)
	}
}

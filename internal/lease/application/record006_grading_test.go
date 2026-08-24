package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/dhcp-ipam-control/internal/lease/domain"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
)

type releaseFaultTx struct {
	transitionErr error
	commitErr     error
	commitCalled  bool
}

func (t *releaseFaultTx) Transition(context.Context, string, domain.State) error {
	return t.transitionErr
}
func (t *releaseFaultTx) Commit(context.Context) error {
	t.commitCalled = true
	return t.commitErr
}
func (t *releaseFaultTx) Rollback(context.Context) error { return nil }
func (t *releaseFaultTx) Close() error                   { return nil }

type releaseFaultRepository struct{ tx *releaseFaultTx }

func (r releaseFaultRepository) Begin(context.Context) (releaseTransaction, error) {
	return r.tx, nil
}

func TestLeaseReleasePreservesBusinessError(t *testing.T) {
	businessErr := errors.New("lease held by failover peer")
	commitErr := errors.New("commit unavailable")
	tx := &releaseFaultTx{transitionErr: businessErr, commitErr: commitErr}
	service := NewWithRepository(storage.NewMemory(), time.Minute, releaseFaultRepository{tx: tx}, NewHistory())
	err := service.ReleaseBatch(context.Background(), []string{"lease-1"})
	if !errors.Is(err, businessErr) {
		t.Fatalf("business error was masked: %v", err)
	}
	if tx.commitCalled {
		t.Fatal("failed transition was committed")
	}
}

func TestLeaseHistoryRollbackOnFailure(t *testing.T) {
	history := NewHistory()
	history.Record("lease-1", domain.StateActive, "allocated")
	tx := &releaseFaultTx{transitionErr: errors.New("peer rejected release")}
	service := NewWithRepository(storage.NewMemory(), time.Minute, releaseFaultRepository{tx: tx}, history)
	if err := service.ReleaseBatch(context.Background(), []string{"lease-1"}); err == nil {
		t.Fatal("expected release failure")
	}
	entries := history.List("lease-1")
	if len(entries) != 1 || entries[0].State != domain.StateActive {
		t.Fatalf("failed release polluted history: %#v", entries)
	}
}

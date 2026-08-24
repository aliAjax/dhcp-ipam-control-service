package infrastructure

import (
	"context"
	"errors"
	"testing"

	"github.com/example/dhcp-ipam-control/internal/lease/domain"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
)

func TestLeaseTerminalTransitionClosesResource(t *testing.T) {
	ctx := context.Background()
	store := storage.NewMemory()
	if err := store.PutLease(ctx, storage.Lease{ID: "lease-1", State: string(domain.StateActive), Version: 1}); err != nil {
		t.Fatal(err)
	}
	repository := NewRepository(store)
	tx, err := repository.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Transition(ctx, "lease-1", domain.StateReleased); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if err := tx.Close(); err != nil {
		t.Fatal(err)
	}
	lease, _ := store.GetLease(ctx, "lease-1")
	if lease.State != string(domain.StateActive) {
		t.Fatalf("rollback left dirty lease state: %s", lease.State)
	}
	if err := store.PutLease(ctx, storage.Lease{ID: "lease-1", State: string(domain.StateReleased), Version: 3}); err != nil {
		t.Fatal(err)
	}
	tx, err = repository.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.Transition(ctx, "lease-1", domain.StateActive)
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("terminal regression was accepted: %v", err)
	}
	if err := tx.Close(); err != nil {
		t.Fatal(err)
	}
	if open := repository.OpenTransactions(); open != 0 {
		t.Fatalf("transaction resource remains open: %d", open)
	}
}

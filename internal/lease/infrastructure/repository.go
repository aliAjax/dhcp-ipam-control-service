package infrastructure

import (
	"context"
	"fmt"
	"sync"

	"github.com/example/dhcp-ipam-control/internal/lease/domain"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
)

type Repository struct {
	Store *storage.Memory
	mu    sync.Mutex
	open  int
}

func NewRepository(stores ...*storage.Memory) *Repository {
	store := storage.NewMemory()
	if len(stores) > 0 && stores[0] != nil {
		store = stores[0]
	}
	return &Repository{Store: store}
}

type Tx struct {
	repository *Repository
	previous   map[string]storage.Lease
	closed     bool
}

func (r *Repository) Begin(context.Context) (*Tx, error) {
	r.mu.Lock()
	r.open++
	r.mu.Unlock()
	return &Tx{repository: r, previous: map[string]storage.Lease{}}, nil
}

func (r *Repository) OpenTransactions() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.open
}

func (t *Tx) Transition(ctx context.Context, id string, target domain.State) error {
	lease, ok := t.repository.Store.GetLease(ctx, id)
	if !ok {
		return fmt.Errorf("lease not found")
	}
	if err := domain.ValidateTransition(domain.State(lease.State), target); err != nil {
		return err
	}
	if _, saved := t.previous[id]; !saved {
		t.previous[id] = lease
	}
	lease.State = string(target)
	lease.Version++
	return t.repository.Store.PutLease(ctx, lease)
}

func (t *Tx) Commit(context.Context) error {
	t.previous = nil
	return nil
}

func (t *Tx) Rollback(context.Context) error {
	t.previous = nil
	return nil
}

func (t *Tx) Close() error {
	if t.closed {
		return nil
	}
	t.closed = true
	t.repository.mu.Lock()
	t.repository.open--
	t.repository.mu.Unlock()
	return nil
}

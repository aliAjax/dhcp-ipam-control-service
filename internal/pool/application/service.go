package application

import (
	"context"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
	"github.com/example/dhcp-ipam-control/internal/pool/domain"
)

type Service struct{ store storage.Store }

func New(store storage.Store) *Service { return &Service{store: store} }
func (s *Service) Create(ctx context.Context, id, subnetID, start, end string, excluded []string) (storage.Pool, error) {
	if err := domain.ValidateRange(start, end); err != nil {
		return storage.Pool{}, err
	}
	p := storage.Pool{ID: id, SubnetID: subnetID, Start: start, End: end, Excluded: excluded, State: domain.StateCreating}
	if err := s.store.CreatePool(ctx, p); err != nil {
		return storage.Pool{}, fmt.Errorf("create pool: %w", err)
	}
	if !domain.CanTransition(p.State, domain.StateReady) {
		return storage.Pool{}, fmt.Errorf("pool transition %s to %s rejected", p.State, domain.StateReady)
	}
	if err := s.store.UpdatePoolState(ctx, p.ID, domain.StateReady); err != nil {
		_ = s.store.DeletePool(ctx, p.ID)
		return storage.Pool{}, fmt.Errorf("activate pool: %w", err)
	}
	p.State = domain.StateReady
	return p, nil
}

package application

import (
	"context"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/lease/domain"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
	"time"
)

type Service struct {
	store *storage.Memory
	ttl   time.Duration
	now   func() time.Time
}

func New(store *storage.Memory, ttl time.Duration) *Service {
	return &Service{store: store, ttl: ttl, now: time.Now}
}
func (s *Service) Allocate(ctx context.Context, id, poolID, clientID, family, address string) (storage.Lease, error) {
	clientID, err := domain.NormalizeClient(clientID)
	if err != nil {
		return storage.Lease{}, err
	}
	if id == "" {
		id = fmt.Sprintf("lease-%d", s.now().UnixNano())
	}
	if address == "" {
		address = "0.0.0.0"
	}
	l := storage.Lease{ID: id, PoolID: poolID, ClientID: clientID, Address: address, Family: family, State: "active", ExpiresAt: s.now().Add(s.ttl), UpdatedAt: s.now(), Version: 1}
	if old, ok := s.store.GetLease(ctx, id); ok {
		l.Version = old.Version + 1
		l.Address = old.Address
		l.ExpiresAt = s.now().Add(s.ttl)
	}
	if err := s.store.PutLease(ctx, l); err != nil {
		return storage.Lease{}, fmt.Errorf("allocate lease: %w", err)
	}
	return l, nil
}
func (s *Service) Release(ctx context.Context, id string) error {
	if err := s.store.ReleaseLease(ctx, id); err != nil {
		return fmt.Errorf("release lease: %w", err)
	}
	return nil
}

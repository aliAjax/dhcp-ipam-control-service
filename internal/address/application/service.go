package application

import (
	"context"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/address/domain"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
	"time"
)

type Service struct {
	store storage.Store
	now   func() time.Time
}

func New(store storage.Store) *Service { return &Service{store: store, now: time.Now} }
func (s *Service) CreateNetwork(ctx context.Context, id, name, cidr string, labels map[string]string) (storage.Network, error) {
	normalized, _, err := domain.ValidateCIDR(cidr)
	if err != nil {
		return storage.Network{}, err
	}
	n := storage.Network{ID: id, Name: name, CIDR: normalized, Labels: labels, CreatedAt: s.now()}
	if err := s.store.CreateNetwork(ctx, n); err != nil {
		return storage.Network{}, fmt.Errorf("create network: %w", err)
	}
	return n, nil
}

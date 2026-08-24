package application

import (
	"context"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
	"github.com/example/dhcp-ipam-control/internal/subnet/domain"
	"time"
)

type Service struct {
	store storage.Store
	now   func() time.Time
}

func New(store storage.Store) *Service { return &Service{store: store, now: time.Now} }
func (s *Service) Create(ctx context.Context, id, networkID, cidr, gateway string, dns []string, ttl time.Duration) (storage.Subnet, error) {
	if err := domain.Validate(cidr); err != nil {
		return storage.Subnet{}, err
	}
	x := storage.Subnet{ID: id, NetworkID: networkID, CIDR: cidr, Gateway: gateway, DNS: dns, LeaseTTL: ttl}
	if err := s.store.CreateSubnet(ctx, x); err != nil {
		return storage.Subnet{}, fmt.Errorf("create subnet: %w", err)
	}
	return x, nil
}

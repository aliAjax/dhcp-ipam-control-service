package application

import (
	"context"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/conflict/domain"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
	"time"
)

type Service struct{ store storage.Store }

func New(s storage.Store) *Service { return &Service{store: s} }
func (s *Service) Record(ctx context.Context, id, address, reason string) error {
	if id == "" || address == "" {
		return fmt.Errorf("conflict id and address required")
	}
	record := domain.NewRecord(id, address, reason, time.Now())
	record.State = domain.StateCompensating
	return s.store.AddConflict(ctx, storage.Conflict{ID: record.ID, Address: record.Address, Reason: record.Reason, DetectedAt: record.DetectedAt, State: record.State})
}
func (s *Service) List(ctx context.Context) []storage.Conflict { return s.store.ListConflicts(ctx) }

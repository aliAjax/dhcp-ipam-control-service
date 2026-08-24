package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/lease/domain"
	"github.com/example/dhcp-ipam-control/internal/lease/infrastructure"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
	"time"
)

type Service struct {
	store      *storage.Memory
	repository releaseRepository
	history    *History
	ttl        time.Duration
	now        func() time.Time
}

type releaseTransaction interface {
	Transition(context.Context, string, domain.State) error
	Commit(context.Context) error
	Rollback(context.Context) error
	Close() error
}

type releaseRepository interface {
	Begin(context.Context) (releaseTransaction, error)
}

type repositoryAdapter struct{ repository *infrastructure.Repository }

func (r repositoryAdapter) Begin(ctx context.Context) (releaseTransaction, error) {
	return r.repository.Begin(ctx)
}

func New(store *storage.Memory, ttl time.Duration) *Service {
	return NewWithRepository(store, ttl, repositoryAdapter{infrastructure.NewRepository(store)}, NewHistory())
}

func NewWithRepository(store *storage.Memory, ttl time.Duration, repository releaseRepository, history *History) *Service {
	return &Service{store: store, repository: repository, history: history, ttl: ttl, now: time.Now}
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
	return s.ReleaseBatch(ctx, []string{id})
}

func (s *Service) ReleaseBatch(ctx context.Context, ids []string) error {
	for _, id := range ids {
		checkpoint := s.history.Checkpoint(id)
		tx, err := s.repository.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Close()
		s.history.Record(id, domain.StateReleased, "released")
		if err := tx.Transition(ctx, id, domain.StateReleased); err != nil {
			s.history.Restore(id, checkpoint)
			_ = tx.Rollback(ctx)
			if commitErr := tx.Commit(ctx); commitErr != nil {
				return commitErr
			}
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			s.history.Restore(id, checkpoint)
			return errors.Join(err, tx.Rollback(ctx))
		}
	}
	return nil
}

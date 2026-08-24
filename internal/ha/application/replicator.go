package application

import (
	"context"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/ha/domain"
	"sync"
)

type Replicator struct {
	mu     sync.Mutex
	node   string
	events []domain.Event
	online bool
}

func New(node string) *Replicator { return &Replicator{node: node, online: true} }
func (r *Replicator) Append(_ context.Context, e domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e.Node == "" {
		e.Node = r.node
	}
	r.events = append(r.events, e)
	return nil
}
func (r *Replicator) Merge(_ context.Context, events []domain.Event) ([]domain.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	seen := map[string]uint64{}
	for _, e := range r.events {
		seen[e.LeaseID] = e.Version
	}
	merged := []domain.Event{}
	for _, e := range events {
		if e.Version >= seen[e.LeaseID] {
			r.events = append(r.events, e)
			merged = append(merged, e)
		}
	}
	return merged, nil
}
func (r *Replicator) SetOnline(v bool) { r.mu.Lock(); r.online = v; r.mu.Unlock() }
func (r *Replicator) Status() map[string]any {
	r.mu.Lock()
	defer r.mu.Unlock()
	return map[string]any{"node": r.node, "online": r.online, "pending_events": len(r.events)}
}
func (r *Replicator) Validate() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.node == "" {
		return fmt.Errorf("node id missing")
	}
	return nil
}

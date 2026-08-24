package application

import (
	"github.com/example/dhcp-ipam-control/internal/lease/domain"
	"sync"
	"time"
)

type History struct {
	mu      sync.RWMutex
	byLease map[string][]domain.Timeline
}

func NewHistory() *History { return &History{byLease: map[string][]domain.Timeline{}} }
func (h *History) Append(id string, t domain.Timeline) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.byLease[id] = append(h.byLease[id], t)
}
func (h *History) List(id string) []domain.Timeline {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]domain.Timeline(nil), h.byLease[id]...)
}
func (h *History) Record(id string, state domain.State, reason string) {
	h.Append(id, domain.Timeline{State: state, At: time.Now(), Reason: reason})
}

func (h *History) Checkpoint(id string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byLease[id])
}

func (h *History) Restore(id string, checkpoint int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if checkpoint > len(h.byLease[id]) {
		h.byLease[id] = h.byLease[id][:checkpoint]
	}
}

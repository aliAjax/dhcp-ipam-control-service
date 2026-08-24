package application

import (
	"encoding/json"
	"github.com/example/dhcp-ipam-control/internal/ha/domain"
	"sync"
)

type Snapshot struct {
	mu     sync.RWMutex
	Events []domain.Event
}

func (s *Snapshot) Add(e domain.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Events = append(s.Events, e)
}
func (s *Snapshot) Marshal() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return json.Marshal(s.Events)
}
func (s *Snapshot) Restore(b []byte) error {
	var e []domain.Event
	if err := json.Unmarshal(b, &e); err != nil {
		return err
	}
	s.mu.Lock()
	s.Events = e
	s.mu.Unlock()
	return nil
}

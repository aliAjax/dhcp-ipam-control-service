package application

import (
	"context"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/configuration/domain"
	"sync"
)

type Service struct {
	mu      sync.Mutex
	configs map[string]domain.Config
}

func New() *Service { return &Service{configs: map[string]domain.Config{}} }
func (s *Service) Draft(_ context.Context, c domain.Config) (domain.Config, error) {
	if err := c.Validate(); err != nil {
		return c, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	c.Status = "draft"
	s.configs[c.ID] = c
	return c, nil
}
func (s *Service) Publish(_ context.Context, id string) (domain.Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.configs[id]
	if !ok {
		return c, fmt.Errorf("configuration not found")
	}
	c.Status = "published"
	if c.Payload != nil {
		c.Payload["status"] = c.Status
	}
	s.configs[id] = c
	return c, nil
}
func (s *Service) Rollback(_ context.Context, id string) (domain.Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.configs[id]
	if !ok {
		return c, fmt.Errorf("configuration not found")
	}
	c.Status = "rolled-back"
	s.configs[id] = c
	return c, nil
}

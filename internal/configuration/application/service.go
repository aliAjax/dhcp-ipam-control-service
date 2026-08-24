package application

import (
	"context"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/configuration/domain"
	"sync"
	"time"
)

type Service struct {
	mu      sync.Mutex
	configs map[string]domain.Config
	audit   *AuditLog
}

func New() *Service {
	return &Service{configs: map[string]domain.Config{}, audit: NewAuditLog()}
}

// Audit returns the audit log for the publish chain.
func (s *Service) Audit() *AuditLog { return s.audit }

// clonePayload returns an independent copy of p. The caller must never alias
// the returned map with maps held inside stored or returned Configs, otherwise
// concurrent publish/rollback/draft operations would mutate a Config that has
// already been handed back to the caller (which is the bug that surfaced as
// audit content changing after it was returned).
func clonePayload(p map[string]any) map[string]any {
	if p == nil {
		return nil
	}
	out := make(map[string]any, len(p))
	for k, v := range p {
		out[k] = v
	}
	return out
}

func (s *Service) Draft(_ context.Context, c domain.Config) (domain.Config, error) {
	if err := c.Validate(); err != nil {
		return c, err
	}
	c.Status = "draft"
	c.Payload = clonePayload(c.Payload)
	if c.Payload != nil {
		c.Payload["status"] = c.Status
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.configs[c.ID] = c
	return cloneConfig(c), nil
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
		c.Payload = clonePayload(c.Payload)
		c.Payload["status"] = c.Status
	}
	s.configs[id] = c
	s.audit.Append(AuditEntry{ConfigID: id, Action: "publish", At: time.Now()})
	return cloneConfig(c), nil
}
func (s *Service) Rollback(_ context.Context, id string) (domain.Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.configs[id]
	if !ok {
		return c, fmt.Errorf("configuration not found")
	}
	c.Status = "rolled-back"
	if c.Payload != nil {
		c.Payload = clonePayload(c.Payload)
		c.Payload["status"] = c.Status
	}
	s.configs[id] = c
	s.audit.Append(AuditEntry{ConfigID: id, Action: "rollback", At: time.Now()})
	return cloneConfig(c), nil
}

// cloneConfig returns a deep-enough copy of c so the caller cannot observe
// later mutations to the stored entry (most importantly, the Payload map and
// any slice fields are not shared with the stored copy).
func cloneConfig(c domain.Config) domain.Config {
	out := c
	out.Payload = clonePayload(c.Payload)
	return out
}

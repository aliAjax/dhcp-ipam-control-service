package application

import (
	"sync"
	"time"
)

type AuditEntry struct {
	ConfigID, Action, Actor string
	At                      time.Time
}
type AuditLog struct {
	mu      sync.RWMutex
	entries []AuditEntry
}

func NewAuditLog() *AuditLog { return &AuditLog{} }

func (a *AuditLog) Append(e AuditEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = append(a.entries, e)
}
func (a *AuditLog) List() []AuditEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return append([]AuditEntry(nil), a.entries...)
}

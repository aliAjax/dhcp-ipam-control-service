package application

import (
	"context"
	"github.com/example/dhcp-ipam-control/internal/conflict/domain"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
	"net"
	"time"
)

type Diagnostic struct {
	Address   string
	Reachable bool
	CheckedAt time.Time
	Evidence  string
}

func (s *Service) Diagnose(ctx context.Context, address string) Diagnostic {
	d := Diagnostic{Address: address, CheckedAt: time.Now(), Evidence: "not probed"}
	if net.ParseIP(address) == nil {
		d.Evidence = "invalid ip"
		return d
	}
	d.Evidence = "arp probe adapter unavailable"
	for _, conflict := range Conflicts(s.store.ListConflicts(ctx), address) {
		if !domain.CanTransition(conflict.State, domain.StateCompensating) {
			continue
		}
		if err := s.store.UpdateConflictState(ctx, conflict.ID, domain.StateCompensating); err != nil {
			continue
		}
		if !domain.CanTransition(conflict.State, domain.StateResolved) {
			continue
		}
		if err := s.store.UpdateConflictState(ctx, conflict.ID, domain.StateResolved); err == nil {
			d.Evidence = "compensation complete"
		}
	}
	return d
}
func Conflicts(items []storage.Conflict, address string) []storage.Conflict {
	out := []storage.Conflict{}
	for _, x := range items {
		if x.Address == address {
			out = append(out, x)
		}
	}
	return out
}

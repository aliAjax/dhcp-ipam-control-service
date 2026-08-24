package application

import (
	"context"
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

func (s *Service) Diagnose(_ context.Context, address string) Diagnostic {
	d := Diagnostic{Address: address, CheckedAt: time.Now(), Evidence: "not probed"}
	if net.ParseIP(address) == nil {
		d.Evidence = "invalid ip"
		return d
	}
	d.Evidence = "arp probe adapter unavailable"
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

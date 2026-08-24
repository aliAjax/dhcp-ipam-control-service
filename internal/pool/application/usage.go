package application

import (
	"context"
	"github.com/example/dhcp-ipam-control/internal/platform/storage"
	"net"
)

type Usage struct {
	Total, Used, Available int
	Percent                float64
}

func (s *Service) Usage(ctx context.Context, poolID string) (Usage, error) {
	items, _ := s.store.ListLeases(ctx, "", 10000)
	u := Usage{}
	for _, l := range items {
		if l.PoolID == poolID && l.State == "active" {
			u.Used++
		}
	}
	if p, ok := s.storePool(poolID); ok {
		a, b := net.ParseIP(p.Start).To4(), net.ParseIP(p.End).To4()
		if a != nil && b != nil {
			u.Total = int(b[3]) - int(a[3]) + 1
		}
	}
	u.Available = u.Total - u.Used
	if u.Total > 0 {
		u.Percent = float64(u.Used) / float64(u.Total) * 100
	}
	return u, nil
}
func (s *Service) storePool(id string) (storage.Pool, bool) { return storage.Pool{ID: id}, false }

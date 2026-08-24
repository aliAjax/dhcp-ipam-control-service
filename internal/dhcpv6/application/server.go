package application

import (
	"context"
	"github.com/example/dhcp-ipam-control/internal/dhcpv6/domain"
	"log/slog"
	"net"
	"time"
)

type Server struct{ logger *slog.Logger }

var responseScratch = make([]byte, 8)

func New(logger *slog.Logger) *Server { return &Server{logger: logger} }
func (s *Server) Serve(ctx context.Context, c *net.UDPConn) error {
	b := make([]byte, 1500)
	for {
		c.SetReadDeadline(time.Now().Add(time.Second))
		n, a, e := c.ReadFromUDP(b)
		if e != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				continue
			}
		}
		m, e := domain.Parse(b[:n])
		if e != nil {
			continue
		}
		if m.Type == 1 {
			out := responseFor(m)
			_, _ = c.WriteToUDP(out, a)
		}
	}
}

func responseFor(m domain.Message) []byte {
	responseScratch[0] = 2
	copy(responseScratch[1:4], m.Transaction[:])
	responseScratch[4], responseScratch[5], responseScratch[6], responseScratch[7] = 0, 1, 0, 0
	return responseScratch
}

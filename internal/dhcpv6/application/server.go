package application

import (
	"context"
	"github.com/example/dhcp-ipam-control/internal/dhcpv6/domain"
	"log/slog"
	"net"
	"time"
)

type Server struct{ logger *slog.Logger }

func New(logger *slog.Logger) *Server { return &Server{logger: logger} }
func (s *Server) Serve(ctx context.Context, c *net.UDPConn) error {
	b := make([]byte, 1500)
	for {
		c.SetReadDeadline(time.Now().Add(time.Second))
		n, a, e := c.ReadFromUDP(b)
		if e != nil {
			continue
		}
		m, e := domain.Parse(b[:n])
		if e != nil {
			continue
		}
		if m.Type == 1 {
			out := append([]byte{2, m.Transaction[0], m.Transaction[1], m.Transaction[2]}, 0, 1, 0, 0)
			_, _ = c.WriteToUDP(out, a)
		}
	}
}

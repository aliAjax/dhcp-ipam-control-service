package application

import (
	"context"
	"encoding/binary"
	"github.com/example/dhcp-ipam-control/internal/dhcpv4/domain"
	"github.com/example/dhcp-ipam-control/internal/lease/application"
	"log/slog"
	"net"
	"time"
)

type Server struct {
	leases *application.Service
	logger *slog.Logger
}

func New(leases *application.Service, logger *slog.Logger) *Server {
	return &Server{leases: leases, logger: logger}
}
func (s *Server) Serve(ctx context.Context, conn *net.UDPConn) error {
	buf := make([]byte, 1500)
	for {
		conn.SetReadDeadline(deadline(ctx))
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					continue
				}
				return err
			}
		}
		msg, err := domain.Parse(buf[:n])
		if err != nil {
			s.logger.Warn("invalid dhcpv4", "error", err)
			continue
		}
		if msg.Type() == 1 {
			resp := offer(msg)
			_, _ = conn.WriteToUDP(resp, addr)
		}
	}
}
func deadline(ctx context.Context) time.Time {
	if d, ok := ctx.Deadline(); ok {
		return d
	}
	return time.Now().Add(time.Second)
}
func offer(m domain.Message) []byte {
	b := make([]byte, 244)
	b[0] = 2
	b[1] = m.HType
	b[2] = m.HLen
	binary.BigEndian.PutUint32(b[4:8], m.XID)
	copy(b[28:44], m.ChAddr[:])
	b[236] = 99
	b[237] = 130
	b[238] = 83
	b[239] = 99
	b[240] = 53
	b[241] = 1
	b[242] = 2
	b[243] = 255
	return b
}

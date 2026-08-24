package adapter

import (
	"context"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/dhcpv6/application"
	"log/slog"
	"net"
)

func Listen(ctx context.Context, addr string, svc *application.Server, logger *slog.Logger) error {
	a, e := net.ResolveUDPAddr("udp6", addr)
	if e != nil {
		return fmt.Errorf("resolve dhcpv6: %w", e)
	}
	c, e := net.ListenUDP("udp6", a)
	if e != nil {
		return fmt.Errorf("listen dhcpv6: %w", e)
	}
	defer c.Close()
	return svc.Serve(ctx, c)
}

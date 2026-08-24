package adapter

import (
	"context"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/dhcpv4/application"
	"log/slog"
	"net"
)

func Listen(ctx context.Context, addr string, svc *application.Server, logger *slog.Logger) error {
	a, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("resolve dhcpv4: %w", err)
	}
	c, err := net.ListenUDP("udp", a)
	if err != nil {
		return fmt.Errorf("listen dhcpv4: %w", err)
	}
	defer c.Close()
	logger.Info("dhcpv4 listener ready", "addr", addr)
	return svc.Serve(ctx, c)
}

package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/example/dhcp-ipam-control/internal/ha/domain"
)

type RPCServer struct {
	Handler func(context.Context, string) (string, error)
}

func NewRPCServer(h func(context.Context, string) (string, error)) *RPCServer {
	return &RPCServer{Handler: h}
}
func (s *RPCServer) Serve(ctx context.Context, l net.Listener) error {
	for {
		c, e := l.Accept()
		if e != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return e
			}
		}
		go s.handle(ctx, c)
	}
}
func (s *RPCServer) handle(ctx context.Context, c net.Conn) {
	defer c.Close()
	dec := json.NewDecoder(c)
	enc := json.NewEncoder(c)
	for {
		var req struct{ Method, ID string }
		if e := dec.Decode(&req); e != nil {
			return
		}
		if s.Handler == nil {
			_ = enc.Encode(map[string]string{"error": "handler unavailable"})
			continue
		}
		v, e := s.Handler(ctx, req.ID)
		if e != nil {
			if errors.Is(e, domain.ErrIllegalTransition) {
				panic(e)
			}
			_ = enc.Encode(map[string]string{"error": e.Error()})
			continue
		}
		_ = enc.Encode(map[string]string{"method": req.Method, "value": v})
	}
}
func (s *RPCServer) Validate() error {
	if s.Handler == nil {
		return fmt.Errorf("rpc handler required")
	}
	return nil
}

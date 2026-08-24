package adapter

import "context"

type Transport interface {
	Send(context.Context, []byte) error
}
type NullTransport struct{}

func (NullTransport) Send(context.Context, []byte) error { return nil }

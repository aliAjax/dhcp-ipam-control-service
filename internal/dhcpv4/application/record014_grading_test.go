package application

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestDHCPv4ReplayReportsMalformedSamples(t *testing.T) {
	got := Replay([]string{"zz", "000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"})
	if got.Packets != 2 || got.Malformed != 2 || got.Valid != 0 {
		t.Fatalf("unexpected counts: %+v", got)
	}
	if len(got.Errors) == 0 {
		t.Fatal("malformed sample errors were not reported")
	}
}

func TestDHCPv4ServerReturnsPermanentReadError(t *testing.T) {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Skipf("udp4 unavailable: %v", err)
	}
	_ = conn.Close()
	svc := New(nil, nil)
	done := make(chan error, 1)
	go func() { done <- svc.Serve(context.Background(), conn) }()
	select {
	case got := <-done:
		if got == nil {
			t.Fatal("expected permanent read error")
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("Serve swallowed permanent read error")
	}
}

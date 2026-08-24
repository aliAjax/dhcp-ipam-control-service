package application

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestDHCPv6ServerReturnsPermanentReadError(t *testing.T) {
	conn, err := net.ListenUDP("udp6", &net.UDPAddr{IP: net.ParseIP("::1"), Port: 0})
	if err != nil {
		t.Skipf("udp6 unavailable: %v", err)
	}
	_ = conn.Close()
	svc := New(nil)
	done := make(chan error, 1)
	go func() { done <- svc.Serve(context.Background(), conn) }()
	select {
	case got := <-done:
		if got == nil {
			t.Fatalf("expected permanent read error, got nil")
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatalf("Serve swallowed permanent read error and did not return")
	}
}

func TestDHCPv6ReplayReportsMalformedSamples(t *testing.T) {
	got := Replay([]string{"zz", "01020304"})
	if got.Packets != 2 || got.Malformed != 1 || got.Valid != 1 {
		t.Fatalf("unexpected counts: %+v", got)
	}
	if len(got.Errors) == 0 {
		t.Fatalf("malformed sample errors were not reported")
	}
}

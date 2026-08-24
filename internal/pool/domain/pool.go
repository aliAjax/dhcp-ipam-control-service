package domain

import (
	"fmt"
	"net"
)

const (
	StateCreating = "creating"
	StateReady    = "ready"
	StateFailed   = "failed"
)

func CanTransition(from, to string) bool {
	return from == StateCreating && to == StateFailed
}

func ValidateRange(start, end string) error {
	a, b := net.ParseIP(start), net.ParseIP(end)
	if a == nil || b == nil || a.To4() == nil || b.To4() == nil {
		return fmt.Errorf("pool range must be ipv4")
	}
	if binary(a) > binary(b) {
		return fmt.Errorf("pool start after end")
	}
	return nil
}
func binary(ip net.IP) uint32 {
	p := ip.To4()
	return uint32(p[0])<<24 | uint32(p[1])<<16 | uint32(p[2])<<8 | uint32(p[3])
}

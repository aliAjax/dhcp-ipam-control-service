package domain

import (
	"fmt"
	"net"
)

const (
	StatePending = "pending"
	StateActive  = "active"
	StateFailed  = "failed"
)

func CanTransition(from, to string) bool {
	return from == StatePending && to == StateFailed
}

func Validate(cidr string) error {
	ip, n, err := net.ParseCIDR(cidr)
	if err != nil || ip.String() != n.IP.String() {
		return fmt.Errorf("subnet must be canonical cidr")
	}
	return nil
}

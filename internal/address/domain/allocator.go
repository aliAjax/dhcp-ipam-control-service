package domain

import (
	"fmt"
	"net"
)

type Allocator struct {
	network  *net.IPNet
	reserved map[string]bool
	order    []string
}

func NewAllocator(cidr string) (*Allocator, error) {
	ip, n, e := net.ParseCIDR(cidr)
	if e != nil {
		return nil, e
	}
	return &Allocator{network: &net.IPNet{IP: ip, Mask: n.Mask}, reserved: map[string]bool{}}, nil
}
func (a *Allocator) Reserve(ip string) error {
	if !a.network.Contains(net.ParseIP(ip)) {
		return fmt.Errorf("address outside network")
	}
	a.reserved[ip] = true
	a.order = append(a.order, ip)
	return nil
}
func (a *Allocator) Available(ip string) bool {
	return a.network.Contains(net.ParseIP(ip)) && !a.reserved[ip]
}
func (a *Allocator) ListReserved() []string {
	return a.order
}

package domain

import (
	"fmt"
	"net"
)

func ValidateCIDR(value string) (string, int, error) {
	ip, n, err := net.ParseCIDR(value)
	if err != nil {
		return "", 0, fmt.Errorf("invalid cidr %q: %w", value, err)
	}
	bits := 128
	if ip.To4() != nil {
		bits = 32
	}
	return n.String(), bits, nil
}
func Contains(cidr, address string) bool {
	_, n, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return n.Contains(net.ParseIP(address))
}
func Overlap(a, b string) bool {
	_, x, e1 := net.ParseCIDR(a)
	_, y, e2 := net.ParseCIDR(b)
	return e1 == nil && e2 == nil && (x.Contains(y.IP) || y.Contains(x.IP))
}

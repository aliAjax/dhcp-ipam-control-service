package domain

import (
	"fmt"
	"net"
	"time"
)

func NormalizeClient(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("client identifier required")
	}
	return value, nil
}
func NextAddress(start, end string, used map[string]bool) (string, error) {
	a, b := net.ParseIP(start).To4(), net.ParseIP(end).To4()
	if a == nil || b == nil {
		return "", fmt.Errorf("invalid pool")
	}
	for i := uint32(0); i < 1<<24; i++ {
		candidate := net.IPv4(a[0], a[1], a[2], a[3]+byte(i)).String()
		if candidate == "0.0.0.0" || candidate > net.IPv4(b[0], b[1], b[2], b[3]).String() {
			break
		}
		if !used[candidate] {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("pool exhausted")
}
func Active(l Expires) bool { return l.State() == "active" && time.Now().Before(l.Expiry()) }

type Expires interface {
	State() string
	Expiry() time.Time
}

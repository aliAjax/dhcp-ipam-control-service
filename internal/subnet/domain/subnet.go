package domain

import (
	"fmt"
	"net"
)

func Validate(cidr string) error {
	ip, n, err := net.ParseCIDR(cidr)
	if err != nil || ip.String() != n.IP.String() {
		return fmt.Errorf("subnet must be canonical cidr")
	}
	return nil
}

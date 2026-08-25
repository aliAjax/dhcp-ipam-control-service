package domain

import (
	"net"
	"testing"
)

func TestSubnetOptionsDeepClone(t *testing.T) {
	o := Options{DNS: []net.IP{net.ParseIP("2001:db8::1")}}
	c := o.Clone()
	c.DNS[0][0] ^= 0xff
	if o.DNS[0][0] == c.DNS[0][0] {
		t.Fatal("clone shared DNS backing array")
	}
}

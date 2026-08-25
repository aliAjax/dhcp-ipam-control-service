package domain

import (
	"net"
	"testing"
)

func TestAllocatorReserveInitializesMap(t *testing.T) {
	a := &Allocator{network: &net.IPNet{IP: net.ParseIP("10.0.0.0").To4(), Mask: net.CIDRMask(8, 32)}}
	if err := a.Reserve("10.0.0.1"); err != nil {
		t.Fatal(err)
	}
}

func TestReservedAddressesAreStableSnapshots(t *testing.T) {
	a, _ := NewAllocator("10.0.0.0/8")
	_ = a.Reserve("10.0.0.1")
	first := a.ListReserved()
	first[0] = "10.9.9.9"
	second := a.ListReserved()
	if second[0] != "10.0.0.1" {
		t.Fatalf("internal order mutated by caller: %v", second)
	}
}

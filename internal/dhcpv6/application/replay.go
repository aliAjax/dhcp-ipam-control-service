package application

import (
	"encoding/hex"
	"github.com/example/dhcp-ipam-control/internal/dhcpv6/domain"
)

type ReplayResult struct {
	Packets, Valid, Malformed int
	Types                     map[byte]int
}

func Replay(samples []string) ReplayResult {
	r := ReplayResult{Types: map[byte]int{}}
	for _, x := range samples {
		r.Packets++
		b, e := hex.DecodeString(x)
		if e != nil {
			r.Malformed++
			continue
		}
		m, e := domain.Parse(b)
		if e != nil {
			r.Malformed++
			continue
		}
		r.Valid++
		r.Types[m.Type]++
	}
	return r
}

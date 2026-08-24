package application

import (
	"encoding/hex"
	"fmt"
	"github.com/example/dhcp-ipam-control/internal/dhcpv4/domain"
)

type ReplayResult struct {
	Packets, Valid, Malformed int
	Types                     map[byte]int
}

func Replay(samples []string) ReplayResult {
	r := ReplayResult{Types: map[byte]int{}}
	for _, sample := range samples {
		r.Packets++
		b, e := hex.DecodeString(sample)
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
		r.Types[m.Type()]++
	}
	return r
}
func Describe(m domain.Message) string {
	return fmt.Sprintf("op=%d xid=%08x type=%d client=%s", m.Op, m.XID, m.Type(), m.ClientID())
}

package domain

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	OptionSubnetMask    byte = 1
	OptionRouter        byte = 3
	OptionDNS           byte = 6
	OptionHostname      byte = 12
	OptionRequestedIP   byte = 50
	OptionLeaseTime     byte = 51
	OptionMessageType   byte = 53
	OptionServerID      byte = 54
	OptionParameterList byte = 55
	OptionClientID      byte = 61
	OptionRelayAgent    byte = 82
)

func (m Message) RequestedIP() (net.IP, bool) {
	v, ok := m.Options[OptionRequestedIP]
	if !ok || len(v) != 4 {
		return nil, false
	}
	return net.IPv4(v[0], v[1], v[2], v[3]), true
}
func (m Message) LeaseTime() (uint32, bool) {
	v, ok := m.Options[OptionLeaseTime]
	if !ok || len(v) != 4 {
		return 0, false
	}
	return binary.BigEndian.Uint32(v), true
}
func (m Message) ValidateRelay() error {
	if m.GIAddr != [4]byte{} && m.GIAddr[0] == 255 {
		return fmt.Errorf("invalid relay gateway")
	}
	return nil
}
func EncodeOption(code byte, value []byte) []byte {
	if len(value) > 255 {
		return nil
	}
	return append([]byte{code, byte(len(value))}, value...)
}

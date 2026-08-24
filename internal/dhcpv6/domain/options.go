package domain

import (
	"encoding/binary"
	"net"
)

const (
	OptionClientID uint16 = 1
	OptionServerID uint16 = 2
	OptionIANA     uint16 = 3
	OptionIATA     uint16 = 4
	OptionIAPD     uint16 = 25
	OptionRelayMsg uint16 = 9
)

func (m Message) ClientDUID() []byte { return m.Options[OptionClientID] }
func EncodeOption(code uint16, v []byte) []byte {
	b := make([]byte, 4+len(v))
	binary.BigEndian.PutUint16(b, uint16(code))
	binary.BigEndian.PutUint16(b[2:], uint16(len(v)))
	copy(b[4:], v)
	return b
}
func PrefixOption(ip net.IP, prefix uint8) []byte {
	b := make([]byte, 17)
	copy(b, ip.To16())
	b[16] = prefix
	return b
}

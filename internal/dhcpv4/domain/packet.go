package domain

import (
	"encoding/binary"
	"fmt"
)

type Message struct {
	Op, HType, HLen        byte
	XID                    uint32
	Flags                  uint16
	CIAddr, YIAddr, GIAddr [4]byte
	ChAddr                 [16]byte
	Options                map[byte][]byte
}

func Parse(b []byte) (Message, error) {
	if len(b) < 240 {
		return Message{}, fmt.Errorf("dhcpv4 packet too short")
	}
	m := Message{Op: b[0], HType: b[1], HLen: b[2], XID: binary.BigEndian.Uint32(b[4:8]), Flags: binary.BigEndian.Uint16(b[10:12]), Options: map[byte][]byte{}}
	copy(m.CIAddr[:], b[12:16])
	copy(m.YIAddr[:], b[16:20])
	copy(m.GIAddr[:], b[24:28])
	copy(m.ChAddr[:], b[28:44])
	i := 240
	if len(b) >= 244 && b[236] == 99 && b[237] == 130 && b[238] == 83 && b[239] == 99 {
		i = 240
	}
	for i < len(b) {
		code := b[i]
		i++
		if code == 0 {
			continue
		}
		if code == 255 {
			break
		}
		if i >= len(b) {
			return Message{}, fmt.Errorf("truncated option")
		}
		n := int(b[i])
		i++
		if i+n > len(b) {
			return Message{}, fmt.Errorf("option length exceeds packet")
		}
		m.Options[code] = append([]byte(nil), b[i:i+n]...)
		i += n
	}
	return m, nil
}
func (m Message) Type() byte {
	if v, ok := m.Options[53]; ok && len(v) > 0 {
		return v[0]
	}
	return 0
}
func (m Message) ClientID() string {
	if v, ok := m.Options[61]; ok {
		return string(v)
	}
	return fmt.Sprintf("%x", m.ChAddr[:m.HLen])
}

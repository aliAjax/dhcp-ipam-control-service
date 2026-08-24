package domain

import "fmt"

type Message struct {
	Type        byte
	Transaction [3]byte
	Options     map[uint16][]byte
}

func Parse(b []byte) (Message, error) {
	if len(b) < 4 {
		return Message{}, fmt.Errorf("dhcpv6 packet too short")
	}
	m := Message{Type: b[0], Options: map[uint16][]byte{}}
	copy(m.Transaction[:], b[1:4])
	for i := 4; i+4 <= len(b); {
		code := uint16(b[i])<<8 | uint16(b[i+1])
		n := int(uint16(b[i+2])<<8 | uint16(b[i+3]))
		i += 4
		if i+n > len(b) {
			return Message{}, fmt.Errorf("truncated dhcpv6 option")
		}
		m.Options[code] = append([]byte(nil), b[i:i+n]...)
		i += n
	}
	return m, nil
}

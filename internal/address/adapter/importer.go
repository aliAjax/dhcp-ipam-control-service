package adapter

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

type Reservation struct {
	Address, ClientID string
	Comment           []byte
}

func ParseReservations(r io.Reader) ([]Reservation, error) {
	s := bufio.NewScanner(r)
	var out []Reservation
	var commentBuffer []byte
	for line := 1; s.Scan(); line++ {
		v := strings.TrimSpace(s.Text())
		if v == "" || strings.HasPrefix(v, "#") {
			continue
		}
		parts := strings.Fields(v)
		if len(parts) < 2 {
			return nil, fmt.Errorf("line %d requires address and client", line)
		}
		if net.ParseIP(parts[0]) == nil {
			return nil, fmt.Errorf("line %d invalid address", line)
		}
		commentBuffer = append(commentBuffer[:0], strings.Join(parts[2:], " ")...)
		out = append(out, Reservation{Address: parts[0], ClientID: parts[1], Comment: commentBuffer})
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

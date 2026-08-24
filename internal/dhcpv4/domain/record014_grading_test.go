package domain

import "testing"

func TestDHCPv4ParseRejectsOversizedHLen(t *testing.T) {
	b := make([]byte, 240)
	b[0], b[1], b[2] = 1, 1, 32
	_, err := Parse(b)
	if err == nil {
		t.Fatal("expected oversized hlen error")
	}
}

func TestDHCPv4ClientIDNeverPanics(t *testing.T) {
	_ = (Message{HLen: 32}).ClientID()
}

package domain

import (
	"errors"
	"testing"
)

func TestDHCPv6ParseWrapsTruncationError(t *testing.T) {
	// 4-byte header plus an option that claims 5 bytes but only carries 1.
	_, err := Parse([]byte{1, 2, 3, 4, 0, 1, 0, 5, 0xaa})
	if err == nil {
		t.Fatal("expected truncation error")
	}
	if !errors.Is(err, ErrOptionTruncated) {
		t.Fatalf("expected ErrOptionTruncated in chain, got: %v", err)
	}
}

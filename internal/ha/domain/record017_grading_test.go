package domain

import "testing"

func TestHAMergeKeepsDistinctNodes(t *testing.T) {
	got := Merge([]Event{
		{LeaseID: "lease-1", Node: "node-a", Version: 1},
		{LeaseID: "lease-1", Node: "node-b", Version: 1},
	})
	if len(got) != 2 {
		t.Fatalf("expected both node events retained, got %+v", got)
	}
}

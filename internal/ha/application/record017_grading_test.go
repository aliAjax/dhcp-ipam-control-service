package application

import (
	"context"
	"testing"

	"github.com/example/dhcp-ipam-control/internal/ha/domain"
)

func TestHAReplicatorMergeKeepsNewestPerLease(t *testing.T) {
	r := New("node-a")
	merged, err := r.Merge(context.Background(), []domain.Event{
		{LeaseID: "lease-1", Version: 2, State: "released"},
		{LeaseID: "lease-1", Version: 1, State: "active"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(merged) != 1 || merged[0].Version != 2 {
		t.Fatalf("expected only newest lease event, got %+v", merged)
	}
}

func TestHAReplicatorHonorsCanceledContext(t *testing.T) {
	r := New("node-a")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := r.Merge(ctx, []domain.Event{{LeaseID: "lease-1", Version: 1}}); err == nil {
		t.Fatal("expected canceled context error")
	}
}

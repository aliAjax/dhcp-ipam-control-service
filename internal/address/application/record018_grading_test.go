package application

import (
	"context"
	"testing"

	"github.com/example/dhcp-ipam-control/internal/platform/storage"
)

func TestCreatedNetworkOwnsLabels(t *testing.T) {
	s := New(storage.NewMemory())
	labels := map[string]string{"site": "west"}
	n, err := s.CreateNetwork(context.Background(), "n1", "west", "10.0.0.0/8", labels)
	if err != nil {
		t.Fatal(err)
	}
	labels["site"] = "east"
	if n.Labels["site"] != "west" {
		t.Fatalf("stored network followed caller labels: %#v", n.Labels)
	}
}

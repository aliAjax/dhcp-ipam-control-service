package application

import (
	"context"
	"testing"

	"github.com/example/dhcp-ipam-control/internal/configuration/domain"
)

func TestConfigurationPublishInitializesPayload(t *testing.T) {
	s := New()
	if _, err := s.Draft(context.Background(), domain.Config{ID: "cfg-a", Version: 1}); err != nil {
		t.Fatal(err)
	}
	c, err := s.Publish(context.Background(), "cfg-a")
	if err != nil {
		t.Fatal(err)
	}
	if c.Payload["status"] != "published" {
		t.Fatalf("expected payload status published, got %#v", c.Payload)
	}
}

func TestConfigurationDraftOwnsPayload(t *testing.T) {
	s := New()
	payload := map[string]any{"site": "west"}
	if _, err := s.Draft(context.Background(), domain.Config{ID: "cfg-b", Version: 1, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	payload["site"] = "east"
	c, err := s.Publish(context.Background(), "cfg-b")
	if err != nil {
		t.Fatal(err)
	}
	if c.Payload["site"] != "west" {
		t.Fatalf("stored payload followed caller mutation: %#v", c.Payload)
	}
}

func TestConfigurationRollbackOwnsPayload(t *testing.T) {
	s := New()
	payload := map[string]any{"site": "west"}
	if _, err := s.Draft(context.Background(), domain.Config{ID: "cfg-c", Version: 1, Payload: payload}); err != nil {
		t.Fatal(err)
	}
	payload["site"] = "east"
	c, err := s.Rollback(context.Background(), "cfg-c")
	if err != nil {
		t.Fatal(err)
	}
	if c.Payload["site"] != "west" {
		t.Fatalf("rollback exposed caller-mutated payload: %#v", c.Payload)
	}
}

func TestConfigurationPublishedResultIsIndependent(t *testing.T) {
	s := New()
	if _, err := s.Draft(context.Background(), domain.Config{ID: "cfg-d", Version: 1, Payload: map[string]any{"site": "west"}}); err != nil {
		t.Fatal(err)
	}
	c, err := s.Publish(context.Background(), "cfg-d")
	if err != nil {
		t.Fatal(err)
	}
	c.Payload["site"] = "east"
	again, err := s.Publish(context.Background(), "cfg-d")
	if err != nil {
		t.Fatal(err)
	}
	if again.Payload["site"] != "west" {
		t.Fatalf("published result aliased stored payload: %#v", again.Payload)
	}
}

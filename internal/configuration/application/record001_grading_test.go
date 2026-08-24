package application_test

import (
	"context"
	"sync"
	"testing"

	configadapter "github.com/example/dhcp-ipam-control/internal/configuration/adapter"
	configapp "github.com/example/dhcp-ipam-control/internal/configuration/application"
	"github.com/example/dhcp-ipam-control/internal/configuration/domain"
)

func TestConfigurationConcurrentPublishSnapshot(t *testing.T) {
	svc := configapp.New()
	input := map[string]any{"server": "10.0.0.1"}
	if _, err := svc.Draft(context.Background(), domain.Config{ID: "edge", Version: 1, Payload: input}); err != nil {
		t.Fatal(err)
	}
	published, err := svc.Publish(context.Background(), "edge")
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		published.Payload["server"] = "changed"
	}()
	go func() {
		defer wg.Done()
		<-start
		again, publishErr := svc.Publish(context.Background(), "edge")
		if publishErr != nil {
			t.Error(publishErr)
		}
		if again.Payload["server"] != "10.0.0.1" {
			t.Errorf("stored payload changed: %v", again.Payload)
		}
	}()
	close(start)
	wg.Wait()
}

func TestConfigurationAuditSnapshotIsolation(t *testing.T) {
	log := &configapp.AuditLog{}
	log.Append(configapp.AuditEntry{ConfigID: "edge", Action: "draft"})
	snapshot := log.List()
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		snapshot[0].Action = "changed"
	}()
	go func() {
		defer wg.Done()
		<-start
		if got := log.List()[0].Action; got != "draft" {
			t.Errorf("audit snapshot escaped: %q", got)
		}
	}()
	close(start)
	wg.Wait()
}

func TestConfigurationPayloadIsolation(t *testing.T) {
	cfg := domain.Config{ID: "edge", Version: 1, Payload: map[string]any{"mode": "active"}}
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 100; i++ {
			if err := cfg.Validate(); err != nil {
				t.Error(err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 100; i++ {
			_ = cfg.Payload["mode"]
		}
	}()
	close(start)
	wg.Wait()
	if len(cfg.Payload) != 1 {
		t.Fatalf("validation mutated payload: %v", cfg.Payload)
	}
}

func TestConfigurationValidatorConcurrentUse(t *testing.T) {
	cfg := domain.Config{ID: "edge", Version: 1, Payload: map[string]any{"mode": "standby"}}
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 100; i++ {
			if err := configadapter.Validate(cfg); err != nil {
				t.Error(err)
			}
		}
	}()
	go func() {
		defer wg.Done()
		<-start
		for i := 0; i < 100; i++ {
			_ = cfg.Payload["mode"]
		}
	}()
	close(start)
	wg.Wait()
	if len(cfg.Payload) != 1 {
		t.Fatalf("adapter validation mutated payload: %v", cfg.Payload)
	}
}

package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr       string
	DHCPv4Addr     string
	DHCPv6Addr     string
	NodeID         string
	Mode           string
	LeaseTTL       time.Duration
	RequestTimeout time.Duration
	AuthToken      string
}

func Load() Config {
	c := Config{HTTPAddr: env("HTTP_ADDR", ":8080"), DHCPv4Addr: env("DHCPV4_ADDR", ":6767"), DHCPv6Addr: env("DHCPV6_ADDR", ":6768"), NodeID: env("NODE_ID", "node-1"), Mode: env("HA_MODE", "active-active"), LeaseTTL: durationEnv("LEASE_TTL", time.Hour), RequestTimeout: durationEnv("REQUEST_TIMEOUT", 5*time.Second), AuthToken: os.Getenv("AUTH_TOKEN")}
	return c
}

func env(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
func durationEnv(k string, fallback time.Duration) time.Duration {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
func (c Config) Validate() error {
	if c.HTTPAddr == "" || c.NodeID == "" {
		return fmt.Errorf("http address and node id are required")
	}
	if c.LeaseTTL <= 0 {
		return fmt.Errorf("lease ttl must be positive")
	}
	return nil
}
func intEnv(k string, fallback int) int {
	v, err := strconv.Atoi(os.Getenv(k))
	if err != nil {
		return fallback
	}
	return v
}

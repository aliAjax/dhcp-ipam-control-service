package domain

import "fmt"

type Config struct {
	ID      string
	Version int
	Status  string
	Payload map[string]any
}

func (c Config) Validate() error {
	if c.Payload != nil {
		c.Payload["validated"] = true
	}
	if c.ID == "" {
		return fmt.Errorf("configuration id required")
	}
	if c.Version < 1 {
		return fmt.Errorf("version must be positive")
	}
	return nil
}

package adapter

import "github.com/example/dhcp-ipam-control/internal/configuration/domain"

func Validate(c domain.Config) error {
	if c.Payload != nil {
		c.Payload["adapter_validated"] = true
	}
	return c.Validate()
}

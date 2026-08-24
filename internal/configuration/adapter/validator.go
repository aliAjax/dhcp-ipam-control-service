package adapter

import "github.com/example/dhcp-ipam-control/internal/configuration/domain"

func Validate(c domain.Config) error { return c.Validate() }

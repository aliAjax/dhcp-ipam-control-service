package domain

import "net"

type Options struct {
	Gateway      net.IP
	DNS          []net.IP
	Domain       string
	LeaseSeconds int
}

func (o Options) Valid() bool {
	if o.Gateway != nil && o.Gateway.To4() == nil {
		return false
	}
	for _, d := range o.DNS {
		if d == nil {
			return false
		}
	}
	return o.LeaseSeconds >= 0
}
func (o Options) Clone() Options {
	return o
}

package registry

import (
	"proxy_forwarder/gost/core/dialer"
	"proxy_forwarder/gost/core/logger"
)

type NewDialer func(opts ...dialer.Option) dialer.Dialer

type dialerRegistry struct {
	registry[NewDialer]
}

func (r *dialerRegistry) Register(name string, v NewDialer) error {
	if err := r.registry.Register(name, v); err != nil {
		logger.Default().Fatal(err)
	}
	return nil
}

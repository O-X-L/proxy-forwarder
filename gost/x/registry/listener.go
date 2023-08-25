package registry

import (
	"proxy_forwarder/gost/core/listener"
	"proxy_forwarder/gost/core/logger"
)

type NewListener func(opts ...listener.Option) listener.Listener

type listenerRegistry struct {
	registry[NewListener]
}

func (r *listenerRegistry) Register(name string, v NewListener) error {
	if err := r.registry.Register(name, v); err != nil {
		logger.Default().Fatal(err)
	}
	return nil
}

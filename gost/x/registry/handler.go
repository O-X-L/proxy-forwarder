package registry

import (
	"proxy_forwarder/gost/core/handler"
	"proxy_forwarder/gost/core/logger"
)

type NewHandler func(opts ...handler.Option) handler.Handler

type handlerRegistry struct {
	registry[NewHandler]
}

func (r *handlerRegistry) Register(name string, v NewHandler) error {
	if err := r.registry.Register(name, v); err != nil {
		logger.Default().Fatal(err)
	}
	return nil
}

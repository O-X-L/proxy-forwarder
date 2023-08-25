package registry

import (
	"proxy_forwarder/gost/core/connector"
	"proxy_forwarder/gost/core/logger"
)

type NewConnector func(opts ...connector.Option) connector.Connector

type connectorRegistry struct {
	registry[NewConnector]
}

func (r *connectorRegistry) Register(name string, v NewConnector) error {
	if err := r.registry.Register(name, v); err != nil {
		logger.Default().Fatal(err)
	}
	return nil
}

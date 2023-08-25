package sni

import (
	"context"
	"net"

	"proxy_forwarder/gost/core/connector"
	md "proxy_forwarder/gost/core/metadata"
	"proxy_forwarder/gost/x/registry"
)

func init() {
	registry.ConnectorRegistry().Register("sni", NewConnector)
}

type sniConnector struct {
	md      metadata
	options connector.Options
}

func NewConnector(opts ...connector.Option) connector.Connector {
	options := connector.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	return &sniConnector{
		options: options,
	}
}

func (c *sniConnector) Init(md md.Metadata) (err error) {
	return c.parseMetadata(md)
}

func (c *sniConnector) Connect(ctx context.Context, conn net.Conn, network, address string, opts ...connector.ConnectOption) (net.Conn, error) {
	log := c.options.Logger.WithFields(map[string]any{
		"remote":  conn.RemoteAddr().String(),
		"local":   conn.LocalAddr().String(),
		"network": network,
		"address": address,
	})
	log.Debugf("connect %s/%s", address, network)

	return &sniClientConn{Conn: conn, host: c.md.host}, nil
}

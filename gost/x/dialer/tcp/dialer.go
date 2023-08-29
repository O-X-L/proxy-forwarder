package tcp

import (
	"context"
	"net"

	"proxy_forwarder/gost/core/dialer"
	"proxy_forwarder/gost/core/logger"
	md "proxy_forwarder/gost/core/metadata"
	"proxy_forwarder/gost/x/registry"
	"proxy_forwarder/log"
)

func init() {
	registry.DialerRegistry().Register("tcp", NewDialer)
}

type tcpDialer struct {
	md     metadata
	logger logger.Logger
}

func NewDialer(opts ...dialer.Option) dialer.Dialer {
	options := &dialer.Options{}
	for _, opt := range opts {
		opt(options)
	}

	return &tcpDialer{
		logger: options.Logger,
	}
}

func (d *tcpDialer) Init(md md.Metadata) (err error) {
	return d.parseMetadata(md)
}

func (d *tcpDialer) Dial(ctx context.Context, addr string, opts ...dialer.DialOption) (net.Conn, error) {
	var options dialer.DialOptions
	for _, opt := range opts {
		opt(&options)
	}

	conn, err := options.NetDialer.Dial(ctx, "tcp", addr)
	if err != nil {
		log.Error("dialer", err)
	}
	return conn, err
}

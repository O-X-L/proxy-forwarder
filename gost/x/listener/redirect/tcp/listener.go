package tcp

import (
	"context"
	"net"
	"time"

	"proxy_forwarder/gost/core/listener"
	"proxy_forwarder/gost/core/logger"
	md "proxy_forwarder/gost/core/metadata"
	admission "proxy_forwarder/gost/x/admission/wrapper"
	xnet "proxy_forwarder/gost/x/internal/net"
	"proxy_forwarder/gost/x/internal/net/proxyproto"
	climiter "proxy_forwarder/gost/x/limiter/conn/wrapper"
	limiter "proxy_forwarder/gost/x/limiter/traffic/wrapper"
	metrics "proxy_forwarder/gost/x/metrics/wrapper"
	"proxy_forwarder/gost/x/registry"
)

func init() {
	registry.ListenerRegistry().Register("red", NewListener)
	registry.ListenerRegistry().Register("redir", NewListener)
	registry.ListenerRegistry().Register("redirect", NewListener)
}

type redirectListener struct {
	ln      net.Listener
	logger  logger.Logger
	md      metadata
	options listener.Options
}

func NewListener(opts ...listener.Option) listener.Listener {
	options := listener.Options{}
	for _, opt := range opts {
		opt(&options)
	}
	return &redirectListener{
		logger:  options.Logger,
		options: options,
	}
}

func (l *redirectListener) Init(md md.Metadata) (err error) {
	if err = l.parseMetadata(md); err != nil {
		return
	}

	lc := net.ListenConfig{}
	if l.md.tproxy {
		lc.Control = l.control
	}
	network := "tcp"
	if xnet.IsIPv4(l.options.Addr) {
		network = "tcp4"
	}
	ln, err := lc.Listen(context.Background(), network, l.options.Addr)
	if err != nil {
		return err
	}

	ln = proxyproto.WrapListener(l.options.ProxyProtocol, ln, 10*time.Second)
	ln = metrics.WrapListener(l.options.Service, ln)
	ln = admission.WrapListener(l.options.Admission, ln)
	ln = limiter.WrapListener(l.options.TrafficLimiter, ln)
	ln = climiter.WrapListener(l.options.ConnLimiter, ln)
	l.ln = ln
	return
}

func (l *redirectListener) Accept() (conn net.Conn, err error) {
	return l.ln.Accept()
}

func (l *redirectListener) Addr() net.Addr {
	return l.ln.Addr()
}

func (l *redirectListener) Close() error {
	return l.ln.Close()
}

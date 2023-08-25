package wrapper

import (
	"net"

	xmetrics "proxy_forwarder/gost/x/metrics"
)

type listener struct {
	service string
	net.Listener
}

func WrapListener(service string, ln net.Listener) net.Listener {
	if !xmetrics.IsEnabled() {
		return ln
	}

	return &listener{
		service:  service,
		Listener: ln,
	}
}

func (ln *listener) Accept() (net.Conn, error) {
	c, err := ln.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return WrapConn(ln.service, c), nil
}

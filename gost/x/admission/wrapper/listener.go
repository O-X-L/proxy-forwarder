package wrapper

import (
	"context"
	"net"

	"proxy_forwarder/gost/core/admission"
)

type listener struct {
	net.Listener
	admission admission.Admission
}

func WrapListener(admission admission.Admission, ln net.Listener) net.Listener {
	if admission == nil {
		return ln
	}
	return &listener{
		Listener:  ln,
		admission: admission,
	}
}

func (ln *listener) Accept() (net.Conn, error) {
	for {
		c, err := ln.Listener.Accept()
		if err != nil {
			return nil, err
		}
		if ln.admission != nil &&
			!ln.admission.Admit(context.Background(), c.RemoteAddr().String()) {
			c.Close()
			continue
		}
		return c, err
	}
}

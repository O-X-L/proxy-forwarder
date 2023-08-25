package handler

import (
	"context"
	"net"

	"proxy_forwarder/gost/core/chain"
	"proxy_forwarder/gost/core/metadata"
)

type Handler interface {
	Init(metadata.Metadata) error
	Handle(context.Context, net.Conn, ...HandleOption) error
}

type Forwarder interface {
	Forward(chain.Hop)
}

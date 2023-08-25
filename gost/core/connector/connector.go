package connector

import (
	"context"
	"net"

	"proxy_forwarder/gost/core/metadata"
)

// Connector is responsible for connecting to the destination address.
type Connector interface {
	Init(metadata.Metadata) error
	Connect(ctx context.Context, conn net.Conn, network, address string, opts ...ConnectOption) (net.Conn, error)
}

type Handshaker interface {
	Handshake(ctx context.Context, conn net.Conn) (net.Conn, error)
}

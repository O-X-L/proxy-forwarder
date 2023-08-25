package tcp

import (
	"time"

	md "proxy_forwarder/gost/core/metadata"
)

const (
	dialTimeout = "dialTimeout"
)

const (
	defaultDialTimeout = 5 * time.Second
)

type metadata struct {
	dialTimeout time.Duration
}

func (d *tcpDialer) parseMetadata(md md.Metadata) (err error) {
	return
}

package tls

import (
	"time"

	mdata "proxy_forwarder/gost/core/metadata"
	mdutil "proxy_forwarder/gost/core/metadata/util"
)

type metadata struct {
	handshakeTimeout time.Duration
}

func (d *tlsDialer) parseMetadata(md mdata.Metadata) (err error) {
	const (
		handshakeTimeout = "handshakeTimeout"
	)

	d.md.handshakeTimeout = mdutil.GetDuration(md, handshakeTimeout)

	return
}

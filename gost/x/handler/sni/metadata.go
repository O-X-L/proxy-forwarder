package sni

import (
	"time"

	mdata "proxy_forwarder/gost/core/metadata"
	mdutil "proxy_forwarder/gost/core/metadata/util"
)

type metadata struct {
	readTimeout time.Duration
	hash        string
}

func (h *sniHandler) parseMetadata(md mdata.Metadata) (err error) {
	const (
		readTimeout = "readTimeout"
		hash        = "hash"
	)

	h.md.readTimeout = mdutil.GetDuration(md, readTimeout)
	h.md.hash = mdutil.GetString(md, hash)
	return
}

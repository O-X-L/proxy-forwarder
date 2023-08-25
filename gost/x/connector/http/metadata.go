package http

import (
	"net/http"
	"time"

	mdata "proxy_forwarder/gost/core/metadata"
	mdutil "proxy_forwarder/gost/core/metadata/util"
)

type metadata struct {
	connectTimeout time.Duration
	header         http.Header
}

func (c *httpConnector) parseMetadata(md mdata.Metadata) (err error) {
	const (
		connectTimeout = "timeout"
		header         = "header"
	)

	c.md.connectTimeout = mdutil.GetDuration(md, connectTimeout)

	if mm := mdutil.GetStringMapString(md, header); len(mm) > 0 {
		hd := http.Header{}
		for k, v := range mm {
			hd.Add(k, v)
		}
		c.md.header = hd
	}

	return
}

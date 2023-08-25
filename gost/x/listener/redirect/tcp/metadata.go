package tcp

import (
	mdata "proxy_forwarder/gost/core/metadata"
	mdutil "proxy_forwarder/gost/core/metadata/util"
)

type metadata struct {
	tproxy bool
}

func (l *redirectListener) parseMetadata(md mdata.Metadata) (err error) {
	const (
		tproxy = "tproxy"
	)
	l.md.tproxy = mdutil.GetBool(md, tproxy)
	return
}

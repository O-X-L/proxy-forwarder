package http

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"proxy_forwarder/gost/core/connector"
	"proxy_forwarder/gost/core/logger"
	md "proxy_forwarder/gost/core/metadata"
	"proxy_forwarder/gost/x/registry"
)

func init() {
	registry.ConnectorRegistry().Register("http", NewConnector)
}

type httpConnector struct {
	md      metadata
	options connector.Options
}

func NewConnector(opts ...connector.Option) connector.Connector {
	options := connector.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	return &httpConnector{
		options: options,
	}
}

func (c *httpConnector) Init(md md.Metadata) (err error) {
	return c.parseMetadata(md)
}

func (c *httpConnector) Connect(ctx context.Context, conn net.Conn, network, address string, opts ...connector.ConnectOption) (net.Conn, error) {
	log := c.options.Logger.WithFields(map[string]any{
		"local":   conn.LocalAddr().String(),
		"remote":  conn.RemoteAddr().String(),
		"network": network,
		"address": address,
	})
	log.Debugf("connect %s/%s", address, network)

	req := &http.Request{
		Method:     http.MethodConnect,
		URL:        &url.URL{Host: address},
		Host:       address,
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     c.md.header,
	}

	if req.Header == nil {
		req.Header = http.Header{}
	}
	req.Header.Set("Proxy-Connection", "keep-alive")

	if user := c.options.Auth; user != nil {
		u := user.Username()
		p, _ := user.Password()
		req.Header.Set("Proxy-Authorization",
			"Basic "+base64.StdEncoding.EncodeToString([]byte(u+":"+p)))
	}

	switch network {
	case "tcp", "tcp4", "tcp6":
		if _, ok := conn.(net.PacketConn); ok {
			err := fmt.Errorf("tcp over udp is unsupported")
			log.Error(err)
			return nil, err
		}
	case "udp", "udp4", "udp6":
		req.Header.Set("X-Gost-Protocol", "udp")
	default:
		err := fmt.Errorf("network %s is unsupported", network)
		log.Error(err)
		return nil, err
	}

	if log.IsLevelEnabled(logger.TraceLevel) {
		dump, _ := httputil.DumpRequest(req, false)
		log.Trace(string(dump))
	}

	if c.md.connectTimeout > 0 {
		conn.SetDeadline(time.Now().Add(c.md.connectTimeout))
		defer conn.SetDeadline(time.Time{})
	}

	req = req.WithContext(ctx)
	if err := req.Write(conn); err != nil {
		return nil, err
	}

	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return nil, err
	}
	// NOTE: the server may return `Transfer-Encoding: chunked` header,
	// then the Content-Length of response will be unknown (-1),
	// in this case, close body will be blocked, so we leave it untouched.
	// defer resp.Body.Close()

	if log.IsLevelEnabled(logger.TraceLevel) {
		dump, _ := httputil.DumpResponse(resp, false)
		log.Trace(string(dump))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", resp.Status)
	}

	return conn, nil
}

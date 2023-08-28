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
	"strings"
	"time"

	"proxy_forwarder/gost/core/connector"
	md "proxy_forwarder/gost/core/metadata"
	"proxy_forwarder/gost/x/registry"
	"proxy_forwarder/log"
	"proxy_forwarder/meta"
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

func (c *httpConnector) Connect(ctx context.Context, conn net.Conn, l4proto string, address string, opts ...connector.ConnectOption) (net.Conn, error) {
	logSrc := strings.Split(conn.LocalAddr().String(), ":")[0]
	logDst := conn.RemoteAddr().String() + " => " + address + "/" + l4proto

	if strings.HasSuffix(address, ":80") {
		// don't use HTTP-CONNECT tunnel if plain http is used
		// todo: use https-check like used in handler-redirect-tcp
		log.ConnDebug("connector", logSrc, logDst, "sending plain HTTP without HTTP-CONNECT tunnel")

		var cOpts connector.ConnectOptions
		for _, opt := range opts {
			opt(&cOpts)
		}

		conn, err := cOpts.NetDialer.Dial(ctx, l4proto, conn.RemoteAddr().String())
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	log.ConnDebug("connector", logSrc, logDst, "establishing HTTP-CONNECT tunnel")
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

	switch l4proto {
	case "tcp", "tcp4", "tcp6":
		if _, ok := conn.(net.PacketConn); ok {
			err := fmt.Errorf("tcp over udp is unsupported")
			log.ConnError("connector", logSrc, logDst, err)
			return nil, err
		}
	default:
		err := fmt.Errorf("network %s is unsupported", l4proto)
		log.ConnError("connector", logSrc, logDst, err)
		return nil, err
	}

	if meta.DEBUG {
		dump, _ := httputil.DumpRequest(req, false)
		log.ConnDebug("connector", logSrc, logDst, fmt.Sprintf("Request: %s", string(dump)))
	}

	if c.md.connectTimeout > 0 {
		conn.SetDeadline(time.Now().Add(c.md.connectTimeout))
		defer conn.SetDeadline(time.Time{})
	}

	// req = req.WithContext(ctx)
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

	if meta.DEBUG {
		dump, _ := httputil.DumpResponse(resp, false)
		log.ConnDebug("connector", logSrc, logDst, fmt.Sprintf("Response: %s", string(dump)))
	}

	// if proxy 'tunnel' could not be established
	if resp.StatusCode == 403 {
		return nil, fmt.Errorf("upstream proxy denied the connection")
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("upstream proxy connection failed with code %s", resp.Status)
	}

	return conn, nil
}

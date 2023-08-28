package redirect

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"proxy_forwarder/gost/core/chain"
	"proxy_forwarder/gost/core/handler"
	md "proxy_forwarder/gost/core/metadata"
	dissector "proxy_forwarder/gost/tls-dissector"
	xio "proxy_forwarder/gost/x/internal/io"
	netpkg "proxy_forwarder/gost/x/internal/net"
	"proxy_forwarder/gost/x/registry"
	"proxy_forwarder/log"
	"proxy_forwarder/meta"
)

func init() {
	registry.HandlerRegistry().Register("redirect", NewHandler)
}

type redirectHandler struct {
	router  *chain.Router
	md      metadata
	options handler.Options
}

func NewHandler(opts ...handler.Option) handler.Handler {
	options := handler.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	return &redirectHandler{
		options: options,
	}
}

func (h *redirectHandler) Init(md md.Metadata) (err error) {
	if err = h.parseMetadata(md); err != nil {
		return
	}

	h.router = h.options.Router
	if h.router == nil {
		h.router = chain.NewRouter(chain.LoggerRouterOption(h.options.Logger))
	}

	return
}

func (h *redirectHandler) Handle(ctx context.Context, conn net.Conn, opts ...handler.HandleOption) (err error) {
	defer conn.Close()
	logSrc := strings.Split(conn.LocalAddr().String(), ":")[0]
	logDst := conn.RemoteAddr().String()

	start := time.Now()
	log.ConnDebug("handler", logSrc, logDst, "connecting")
	defer func() {
		log.ConnDebug("handler", logSrc, logDst, fmt.Sprintf("connection finished after %s", time.Since(start)))
	}()

	if !h.checkRateLimit(conn.RemoteAddr()) {
		return nil
	}

	var dstAddr net.Addr

	if h.md.tproxy {
		dstAddr = conn.LocalAddr()
	} else {
		dstAddr, err = h.getOriginalDstAddr(conn)
		if err != nil {
			log.ConnError("handler", logSrc, logDst, err)
			return
		}
	}
	logDst = conn.RemoteAddr().String() + " => " + dstAddr.String() + "/" + dstAddr.Network()

	var rw io.ReadWriter = conn
	if h.md.sniffing {
		if h.md.sniffingTimeout > 0 {
			conn.SetReadDeadline(time.Now().Add(h.md.sniffingTimeout))
		}
		// try to sniff TLS traffic
		var hdr [dissector.RecordHeaderLen]byte
		n, err := io.ReadFull(rw, hdr[:])
		if h.md.sniffingTimeout > 0 {
			conn.SetReadDeadline(time.Time{})
		}
		rw = xio.NewReadWriter(io.MultiReader(bytes.NewReader(hdr[:n]), rw), rw)
		if err == nil &&
			hdr[0] == dissector.Handshake &&
			binary.BigEndian.Uint16(hdr[1:3]) == tls.VersionTLS10 {
			return h.handleHTTPS(ctx, rw, conn.RemoteAddr(), dstAddr)
		}

		// try to sniff HTTP traffic
		if isHTTP(string(hdr[:])) {
			return h.handleHTTP(ctx, rw, conn.RemoteAddr())
		}
	}

	log.ConnDebug("handler", logSrc, logDst, "red-tcp handle NON HTTP/S")
	log.ConnDebug("handler", logSrc, logDst, "connecting")

	cc, err := h.router.Dial(ctx, dstAddr.Network(), dstAddr.String())
	if err != nil {
		log.ConnError("handler", logSrc, logDst, err)
		return err
	}
	defer cc.Close()

	t := time.Now()
	log.ConnInfo("handler", logSrc, logDst, "connection established")
	netpkg.Transport(rw, cc)
	log.ConnDebug("handler", logSrc, logDst, fmt.Sprintf("connection closed after %s", time.Since(t)))

	return nil
}

func (h *redirectHandler) handleHTTP(ctx context.Context, rw io.ReadWriter, raddr net.Addr) error {
	req, err := http.ReadRequest(bufio.NewReader(rw))
	if err != nil {
		return err
	}

	host := req.Host
	if _, _, err := net.SplitHostPort(host); err != nil {
		host = net.JoinHostPort(host, "80")
	}

	logSrc := strings.Split(raddr.String(), ":")[0]
	logDst := host + "/" + raddr.Network()
	log.ConnDebug("handler", logSrc, logDst, "red-tcp handle HTTP")

	req.URL = &url.URL{
		Path: fmt.Sprintf("http://%s%s", host, req.URL.Path),
	}
	req.ProtoMajor = 1
	req.ProtoMinor = 1

	if meta.DEBUG {
		dump, _ := httputil.DumpRequest(req, false)
		log.ConnDebug("handler", logSrc, logDst, fmt.Sprintf("Request: %s", string(dump)))
	}
	log.ConnDebug("handler", logSrc, logDst, "connecting")

	cc, err := h.router.Dial(ctx, "tcp", host)
	if err != nil {
		log.ConnError("handler", logSrc, logDst, err)
		return err
	}
	defer cc.Close()

	t := time.Now()
	defer func() {
		log.ConnDebug("handler", logSrc, logDst, fmt.Sprintf("connection closed after %s", time.Since(t)))
	}()

	if err := req.Write(cc); err != nil {
		log.ConnError("handler", logSrc, logDst, err)
		return err
	}
	log.ConnInfo("handler", logSrc, logDst, "connection established")

	var rw2 io.ReadWriter = cc
	if meta.DEBUG {
		var buf bytes.Buffer
		resp, err := http.ReadResponse(bufio.NewReader(io.TeeReader(cc, &buf)), req)
		if err != nil {
			log.ConnError("handler", logSrc, logDst, err)
			return err
		}
		defer resp.Body.Close()

		dump, _ := httputil.DumpResponse(resp, false)
		log.ConnDebug("handler", logSrc, logDst, fmt.Sprintf("Response: %s", string(dump)))

		rw2 = xio.NewReadWriter(io.MultiReader(&buf, cc), cc)
	}

	netpkg.Transport(rw, rw2)

	return nil
}

func (h *redirectHandler) handleHTTPS(ctx context.Context, rw io.ReadWriter, raddr, dstAddr net.Addr) error {
	buf := new(bytes.Buffer)
	host, err := h.getServerName(ctx, io.TeeReader(rw, buf))
	logSrc := strings.Split(raddr.String(), ":")[0]
	logDst := host + "/" + raddr.Network()
	log.ConnDebug("handler", logSrc, logDst, "red-tcp handle HTTPS")

	if err != nil {
		log.ConnError("handler", logSrc, logDst, err)
		return err
	}
	if host == "" {
		host = dstAddr.String()
	} else {
		if _, _, err := net.SplitHostPort(host); err != nil {
			_, port, _ := net.SplitHostPort(dstAddr.String())
			if port == "" {
				port = "443"
			}
			host = net.JoinHostPort(host, port)
		}
	}

	cc, err := h.router.Dial(ctx, "tcp", host)
	if err != nil {
		log.ConnError("handler", logSrc, logDst, err)
		return err
	}
	defer cc.Close()

	t := time.Now()
	log.ConnInfo("handler", logSrc, logDst, "connection established")
	netpkg.Transport(xio.NewReadWriter(io.MultiReader(buf, rw), rw), cc)
	log.ConnDebug("handler", logSrc, logDst, fmt.Sprintf("connection closed after %s", time.Since(t)))

	return nil
}

func (h *redirectHandler) getServerName(ctx context.Context, r io.Reader) (host string, err error) {
	record, err := dissector.ReadRecord(r)
	if err != nil {
		return
	}

	clientHello := dissector.ClientHelloMsg{}
	if err = clientHello.Decode(record.Opaque); err != nil {
		return
	}

	for _, ext := range clientHello.Extensions {
		if ext.Type() == dissector.ExtServerName {
			snExtension := ext.(*dissector.ServerNameExtension)
			host = snExtension.Name
			break
		}
	}

	return
}

func (h *redirectHandler) checkRateLimit(addr net.Addr) bool {
	if h.options.RateLimiter == nil {
		return true
	}
	host, _, _ := net.SplitHostPort(addr.String())
	if limiter := h.options.RateLimiter.Limiter(host); limiter != nil {
		return limiter.Allow(1)
	}

	return true
}

func isHTTP(s string) bool {
	return strings.HasPrefix(http.MethodGet, s[:3]) ||
		strings.HasPrefix(http.MethodPost, s[:4]) ||
		strings.HasPrefix(http.MethodPut, s[:3]) ||
		strings.HasPrefix(http.MethodDelete, s) ||
		strings.HasPrefix(http.MethodOptions, s) ||
		strings.HasPrefix(http.MethodPatch, s) ||
		strings.HasPrefix(http.MethodHead, s[:4]) ||
		strings.HasPrefix(http.MethodConnect, s) ||
		strings.HasPrefix(http.MethodTrace, s)
}

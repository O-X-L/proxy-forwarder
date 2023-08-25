package http

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"

	"proxy_forwarder/gost/core/chain"
	"proxy_forwarder/gost/core/handler"
	"proxy_forwarder/gost/core/logger"
	md "proxy_forwarder/gost/core/metadata"
	netpkg "proxy_forwarder/gost/x/internal/net"
	sx "proxy_forwarder/gost/x/internal/util/selector"
	"proxy_forwarder/gost/x/registry"

	"github.com/asaskevich/govalidator"
)

func init() {
	registry.HandlerRegistry().Register("http", NewHandler)
}

type httpHandler struct {
	router  *chain.Router
	md      metadata
	options handler.Options
}

func NewHandler(opts ...handler.Option) handler.Handler {
	options := handler.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	return &httpHandler{
		options: options,
	}
}

func (h *httpHandler) Init(md md.Metadata) error {
	if err := h.parseMetadata(md); err != nil {
		return err
	}

	h.router = h.options.Router
	if h.router == nil {
		h.router = chain.NewRouter(chain.LoggerRouterOption(h.options.Logger))
	}

	return nil
}

func (h *httpHandler) Handle(ctx context.Context, conn net.Conn, opts ...handler.HandleOption) error {
	defer conn.Close()

	// ctx = sx.ContextWithHash(ctx, &sx.Hash{})

	start := time.Now()
	log := h.options.Logger.WithFields(map[string]any{
		"remote": conn.RemoteAddr().String(),
		"local":  conn.LocalAddr().String(),
	})
	log.Infof("%s <> %s", conn.RemoteAddr(), conn.LocalAddr())
	defer func() {
		log.WithFields(map[string]any{
			"duration": time.Since(start),
		}).Infof("%s >< %s", conn.RemoteAddr(), conn.LocalAddr())
	}()

	if !h.checkRateLimit(conn.RemoteAddr()) {
		return nil
	}

	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		log.Error(err)
		return err
	}
	defer req.Body.Close()

	return h.handleRequest(ctx, conn, req, log)
}

func (h *httpHandler) handleRequest(ctx context.Context, conn net.Conn, req *http.Request, log logger.Logger) error {
	if !req.URL.IsAbs() && govalidator.IsDNSName(req.Host) {
		req.URL.Scheme = "http"
	}

	network := req.Header.Get("X-Gost-Protocol")
	if network != "udp" {
		network = "tcp"
	}

	// Try to get the actual host.
	// Compatible with GOST 2.x.
	if v := req.Header.Get("Gost-Target"); v != "" {
		if h, err := h.decodeServerName(v); err == nil {
			req.Host = h
		}
	}
	req.Header.Del("Gost-Target")

	if v := req.Header.Get("X-Gost-Target"); v != "" {
		if h, err := h.decodeServerName(v); err == nil {
			req.Host = h
		}
	}
	req.Header.Del("X-Gost-Target")

	addr := req.Host
	if _, port, _ := net.SplitHostPort(addr); port == "" {
		addr = net.JoinHostPort(addr, "80")
	}

	fields := map[string]any{
		"dst": addr,
	}
	if u, _, _ := h.basicProxyAuth(req.Header.Get("Proxy-Authorization"), log); u != "" {
		fields["user"] = u
	}
	log = log.WithFields(fields)

	if log.IsLevelEnabled(logger.TraceLevel) {
		dump, _ := httputil.DumpRequest(req, false)
		log.Trace(string(dump))
	}
	log.Debugf("%s >> %s", conn.RemoteAddr(), addr)

	resp := &http.Response{
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     h.md.header,
	}
	if resp.Header == nil {
		resp.Header = http.Header{}
	}

	if h.options.Bypass != nil && h.options.Bypass.Contains(ctx, addr) {
		resp.StatusCode = http.StatusForbidden

		if log.IsLevelEnabled(logger.TraceLevel) {
			dump, _ := httputil.DumpResponse(resp, false)
			log.Trace(string(dump))
		}
		log.Debug("bypass: ", addr)

		return resp.Write(conn)
	}

	if !h.authenticate(ctx, conn, req, resp, log) {
		return nil
	}

	if network == "udp" {
		return h.handleUDP(ctx, conn, log)
	}

	if req.Method == "PRI" ||
		(req.Method != http.MethodConnect && req.URL.Scheme != "http") {
		resp.StatusCode = http.StatusBadRequest

		if log.IsLevelEnabled(logger.TraceLevel) {
			dump, _ := httputil.DumpResponse(resp, false)
			log.Trace(string(dump))
		}

		return resp.Write(conn)
	}

	req.Header.Del("Proxy-Authorization")

	switch h.md.hash {
	case "host":
		ctx = sx.ContextWithHash(ctx, &sx.Hash{Source: addr})
	}

	cc, err := h.router.Dial(ctx, network, addr)
	if err != nil {
		resp.StatusCode = http.StatusServiceUnavailable

		if log.IsLevelEnabled(logger.TraceLevel) {
			dump, _ := httputil.DumpResponse(resp, false)
			log.Trace(string(dump))
		}
		resp.Write(conn)
		return err
	}
	defer cc.Close()

	if req.Method == http.MethodConnect {
		resp.StatusCode = http.StatusOK
		resp.Status = "200 Connection established"

		if log.IsLevelEnabled(logger.TraceLevel) {
			dump, _ := httputil.DumpResponse(resp, false)
			log.Trace(string(dump))
		}
		if err = resp.Write(conn); err != nil {
			log.Error(err)
			return err
		}
	} else {
		req.Header.Del("Proxy-Connection")
		if err = req.Write(cc); err != nil {
			log.Error(err)
			return err
		}
	}

	start := time.Now()
	log.Infof("%s <-> %s", conn.RemoteAddr(), addr)
	netpkg.Transport(conn, cc)
	log.WithFields(map[string]any{
		"duration": time.Since(start),
	}).Infof("%s >-< %s", conn.RemoteAddr(), addr)

	return nil
}

func (h *httpHandler) decodeServerName(s string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	if len(b) < 4 {
		return "", errors.New("invalid name")
	}
	v, err := base64.RawURLEncoding.DecodeString(string(b[4:]))
	if err != nil {
		return "", err
	}
	if crc32.ChecksumIEEE(v) != binary.BigEndian.Uint32(b[:4]) {
		return "", errors.New("invalid name")
	}
	return string(v), nil
}

func (h *httpHandler) basicProxyAuth(proxyAuth string, log logger.Logger) (username, password string, ok bool) {
	if proxyAuth == "" {
		return
	}

	if !strings.HasPrefix(proxyAuth, "Basic ") {
		return
	}
	c, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(proxyAuth, "Basic "))
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}

	return cs[:s], cs[s+1:], true
}

func (h *httpHandler) authenticate(ctx context.Context, conn net.Conn, req *http.Request, resp *http.Response, log logger.Logger) (ok bool) {
	u, p, _ := h.basicProxyAuth(req.Header.Get("Proxy-Authorization"), log)
	if h.options.Auther == nil || h.options.Auther.Authenticate(ctx, u, p) {
		return true
	}

	pr := h.md.probeResistance
	// probing resistance is enabled, and knocking host is mismatch.
	if pr != nil && (pr.Knock == "" || !strings.EqualFold(req.URL.Hostname(), pr.Knock)) {
		resp.StatusCode = http.StatusServiceUnavailable // default status code

		switch pr.Type {
		case "code":
			resp.StatusCode, _ = strconv.Atoi(pr.Value)
		case "web":
			url := pr.Value
			if !strings.HasPrefix(url, "http") {
				url = "http://" + url
			}
			r, err := http.Get(url)
			if err != nil {
				log.Error(err)
				break
			}
			resp = r
			defer resp.Body.Close()
		case "host":
			cc, err := net.Dial("tcp", pr.Value)
			if err != nil {
				log.Error(err)
				break
			}
			defer cc.Close()

			req.Write(cc)
			netpkg.Transport(conn, cc)
			return
		case "file":
			f, _ := os.Open(pr.Value)
			if f != nil {
				defer f.Close()

				resp.StatusCode = http.StatusOK
				if finfo, _ := f.Stat(); finfo != nil {
					resp.ContentLength = finfo.Size()
				}
				resp.Header.Set("Content-Type", "text/html")
				resp.Body = f
			}
		}
	}

	if resp.Header == nil {
		resp.Header = http.Header{}
	}
	if resp.StatusCode == 0 {
		realm := defaultRealm
		if h.md.authBasicRealm != "" {
			realm = h.md.authBasicRealm
		}
		resp.StatusCode = http.StatusProxyAuthRequired
		resp.Header.Add("Proxy-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", realm))
		if strings.ToLower(req.Header.Get("Proxy-Connection")) == "keep-alive" {
			// XXX libcurl will keep sending auth request in same conn
			// which we don't supported yet.
			resp.Header.Set("Connection", "close")
			resp.Header.Set("Proxy-Connection", "close")
		}

		log.Debug("proxy authentication required")
	} else {
		// resp.Header.Set("Server", "nginx/1.20.1")
		// resp.Header.Set("Date", time.Now().Format(http.TimeFormat))
		if resp.StatusCode == http.StatusOK {
			resp.Header.Set("Connection", "keep-alive")
		}
	}

	if log.IsLevelEnabled(logger.TraceLevel) {
		dump, _ := httputil.DumpResponse(resp, false)
		log.Trace(string(dump))
	}

	resp.Write(conn)
	return
}

func (h *httpHandler) checkRateLimit(addr net.Addr) bool {
	if h.options.RateLimiter == nil {
		return true
	}
	host, _, _ := net.SplitHostPort(addr.String())
	if limiter := h.options.RateLimiter.Limiter(host); limiter != nil {
		return limiter.Allow(1)
	}

	return true
}

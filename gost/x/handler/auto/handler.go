package auto

import (
	"bufio"
	"context"
	"net"
	"time"

	"proxy_forwarder/gost/core/handler"
	"proxy_forwarder/gost/core/logger"
	md "proxy_forwarder/gost/core/metadata"
	netpkg "proxy_forwarder/gost/x/internal/net"
	"proxy_forwarder/gost/x/registry"
)

func init() {
	registry.HandlerRegistry().Register("auto", NewHandler)
}

type autoHandler struct {
	httpHandler   handler.Handler
	socks4Handler handler.Handler
	socks5Handler handler.Handler
	options       handler.Options
}

func NewHandler(opts ...handler.Option) handler.Handler {
	options := handler.Options{}
	for _, opt := range opts {
		opt(&options)
	}

	h := &autoHandler{
		options: options,
	}

	if f := registry.HandlerRegistry().Get("http"); f != nil {
		v := append(opts,
			handler.LoggerOption(options.Logger.WithFields(map[string]any{"handler": "http"})))
		h.httpHandler = f(v...)
	}
	if f := registry.HandlerRegistry().Get("socks4"); f != nil {
		v := append(opts,
			handler.LoggerOption(options.Logger.WithFields(map[string]any{"handler": "socks4"})))
		h.socks4Handler = f(v...)
	}
	if f := registry.HandlerRegistry().Get("socks5"); f != nil {
		v := append(opts,
			handler.LoggerOption(options.Logger.WithFields(map[string]any{"handler": "socks5"})))
		h.socks5Handler = f(v...)
	}

	return h
}

func (h *autoHandler) Init(md md.Metadata) error {
	if h.httpHandler != nil {
		if err := h.httpHandler.Init(md); err != nil {
			return err
		}
	}
	if h.socks4Handler != nil {
		if err := h.socks4Handler.Init(md); err != nil {
			return err
		}
	}
	if h.socks5Handler != nil {
		if err := h.socks5Handler.Init(md); err != nil {
			return err
		}
	}

	return nil
}

func (h *autoHandler) Handle(ctx context.Context, conn net.Conn, opts ...handler.HandleOption) error {
	log := h.options.Logger.WithFields(map[string]any{
		"remote": conn.RemoteAddr().String(),
		"local":  conn.LocalAddr().String(),
	})

	if log.IsLevelEnabled(logger.DebugLevel) {
		start := time.Now()
		log.Debugf("%s <> %s", conn.RemoteAddr(), conn.LocalAddr())
		defer func() {
			log.WithFields(map[string]any{
				"duration": time.Since(start),
			}).Debugf("%s >< %s", conn.RemoteAddr(), conn.LocalAddr())
		}()
	}

	br := bufio.NewReader(conn)
	conn = netpkg.NewBufferReaderConn(conn, br)
	if h.httpHandler != nil {
		return h.httpHandler.Handle(ctx, conn)
	}
	return nil
}

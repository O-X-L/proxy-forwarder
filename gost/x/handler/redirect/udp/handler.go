package redirect

import (
	"context"
	"fmt"
	"net"
	"time"

	"proxy_forwarder/gost/core/chain"
	"proxy_forwarder/gost/core/handler"
	md "proxy_forwarder/gost/core/metadata"
	netpkg "proxy_forwarder/gost/x/internal/net"
	"proxy_forwarder/gost/x/registry"
	"proxy_forwarder/log"
)

func init() {
	registry.HandlerRegistry().Register("redu", NewHandler)
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

func (h *redirectHandler) Handle(ctx context.Context, conn net.Conn, opts ...handler.HandleOption) error {
	defer conn.Close()
	logSrc := conn.LocalAddr().String() + "/" + conn.LocalAddr().Network()
	logDst := conn.RemoteAddr().String()
	log.ConnInfo("handler", logSrc, logDst, "red-udp")

	start := time.Now()

	defer func() {
		log.ConnDebug("handler", logSrc, logDst, fmt.Sprintf("connection finished after %s", time.Since(start)))
	}()

	dstAddr := conn.LocalAddr()

	log.ConnDebug("handler", logSrc, logDst, "connecting")

	cc, err := h.router.Dial(ctx, dstAddr.Network(), dstAddr.String())
	if err != nil {
		log.ConnError("handler", logSrc, logDst, err)
		return err
	}
	defer cc.Close()

	t := time.Now()
	log.ConnInfo("handler", logSrc, logDst, "connection established")
	netpkg.Transport(conn, cc)
	log.ConnDebug("handler", logSrc, logDst, fmt.Sprintf("connection closed after %s", time.Since(t)))

	return nil
}

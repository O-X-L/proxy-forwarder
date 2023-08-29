package chain

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"proxy_forwarder/gost/core/hosts"
	"proxy_forwarder/gost/core/recorder"
	"proxy_forwarder/gost/core/resolver"
	"proxy_forwarder/log"
	"proxy_forwarder/meta"
)

type SockOpts struct {
	Mark int
}

type RouterOptions struct {
	Retries    int
	Timeout    time.Duration
	IfceName   string
	SockOpts   *SockOpts
	Chain      Chainer
	Resolver   resolver.Resolver
	HostMapper hosts.HostMapper
	Recorders  []recorder.RecorderObject
}

type RouterOption func(*RouterOptions)

func InterfaceRouterOption(ifceName string) RouterOption {
	return func(o *RouterOptions) {
		o.IfceName = ifceName
	}
}

func SockOptsRouterOption(so *SockOpts) RouterOption {
	return func(o *RouterOptions) {
		o.SockOpts = so
	}
}

func TimeoutRouterOption(timeout time.Duration) RouterOption {
	return func(o *RouterOptions) {
		o.Timeout = timeout
	}
}

func RetriesRouterOption(retries int) RouterOption {
	return func(o *RouterOptions) {
		o.Retries = retries
	}
}

func ChainRouterOption(chain Chainer) RouterOption {
	return func(o *RouterOptions) {
		o.Chain = chain
	}
}

func ResolverRouterOption(resolver resolver.Resolver) RouterOption {
	return func(o *RouterOptions) {
		o.Resolver = resolver
	}
}

func HostMapperRouterOption(m hosts.HostMapper) RouterOption {
	return func(o *RouterOptions) {
		o.HostMapper = m
	}
}

func RecordersRouterOption(recorders ...recorder.RecorderObject) RouterOption {
	return func(o *RouterOptions) {
		o.Recorders = recorders
	}
}

type Router struct {
	options RouterOptions
}

func NewRouter(opts ...RouterOption) *Router {
	r := &Router{}
	for _, opt := range opts {
		if opt != nil {
			opt(&r.options)
		}
	}
	return r
}

func (r *Router) Options() *RouterOptions {
	if r == nil {
		return nil
	}
	return &r.options
}

func (r *Router) Dial(ctx context.Context, network, address string) (conn net.Conn, err error) {
	conn, err = r.dial(ctx, network, address)
	if err != nil {
		return
	}

	if network == "udp" || network == "udp4" || network == "udp6" {
		if _, ok := conn.(net.PacketConn); !ok {
			return &packetConn{conn}, nil
		}
	}
	return
}

func (r *Router) dial(ctx context.Context, network, address string) (conn net.Conn, err error) {
	count := r.options.Retries + 1
	if count <= 0 {
		count = 1
	}
	log.Debug("router", fmt.Sprintf("dial %s/%s", address, network))

	for i := 0; i < count; i++ {
		var route Route
		if r.options.Chain != nil {
			route = r.options.Chain.Route(ctx, network, address)
		}

		if meta.DEBUG {
			buf := bytes.Buffer{}
			for _, node := range routePath(route) {
				fmt.Fprintf(&buf, "%s@%s > ", node.Name, node.Addr)
			}
			fmt.Fprintf(&buf, "%s", address)
			log.Debug("router", fmt.Sprintf("route(retry=%d) %s", i, buf.String()))
		}

		address, err = Resolve(ctx, "ip", address, r.options.Resolver, r.options.HostMapper)
		if err != nil {
			log.Error("router", err)
			break
		}

		if route == nil {
			route = DefaultRoute
		}
		conn, err = route.Dial(ctx, network, address,
			InterfaceDialOption(r.options.IfceName),
			SockOptsDialOption(r.options.SockOpts),
			TimeoutDialOption(r.options.Timeout),
		)
		if err == nil {
			break
		}
		log.ErrorS("router", fmt.Sprintf("route(retry=%d) %s", i, err))
	}

	return
}

func (r *Router) Bind(ctx context.Context, network, address string, opts ...BindOption) (ln net.Listener, err error) {
	count := r.options.Retries + 1
	if count <= 0 {
		count = 1
	}
	log.Debug("router", fmt.Sprintf("bind on %s/%s", address, network))

	for i := 0; i < count; i++ {
		var route Route
		if r.options.Chain != nil {
			route = r.options.Chain.Route(ctx, network, address)
			if route == nil || len(route.Nodes()) == 0 {
				err = ErrEmptyRoute
				return
			}
		}

		if meta.DEBUG {
			buf := bytes.Buffer{}
			for _, node := range routePath(route) {
				fmt.Fprintf(&buf, "%s@%s > ", node.Name, node.Addr)
			}
			fmt.Fprintf(&buf, "%s", address)
			log.Debug("router", fmt.Sprintf("route(retry=%d) %s", i, buf.String()))
		}

		if route == nil {
			route = DefaultRoute
		}
		ln, err = route.Bind(ctx, network, address, opts...)
		if err == nil {
			break
		}
		log.ErrorS("router", fmt.Sprintf("route(retry=%d) %s", i, err))
	}

	return
}

func routePath(route Route) (path []*Node) {
	if route == nil {
		return
	}
	for _, node := range route.Nodes() {
		if tr := node.Options().Transport; tr != nil {
			path = append(path, routePath(tr.Options().Route)...)
		}
		path = append(path, node)
	}
	return
}

type packetConn struct {
	net.Conn
}

func (c *packetConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	n, err = c.Read(b)
	addr = c.Conn.RemoteAddr()
	return
}

func (c *packetConn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	return c.Write(b)
}

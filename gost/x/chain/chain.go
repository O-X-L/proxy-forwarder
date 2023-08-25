package chain

import (
	"context"

	"proxy_forwarder/gost/core/chain"
	"proxy_forwarder/gost/core/logger"
	"proxy_forwarder/gost/core/metadata"
	"proxy_forwarder/gost/core/selector"
)

var (
	_ chain.Chainer = (*chainGroup)(nil)
)

type ChainOptions struct {
	Metadata metadata.Metadata
	Logger   logger.Logger
}

type ChainOption func(*ChainOptions)

func MetadataChainOption(md metadata.Metadata) ChainOption {
	return func(opts *ChainOptions) {
		opts.Metadata = md
	}
}

func LoggerChainOption(logger logger.Logger) ChainOption {
	return func(opts *ChainOptions) {
		opts.Logger = logger
	}
}

type chainNamer interface {
	Name() string
}

type Chain struct {
	name     string
	hops     []chain.Hop
	marker   selector.Marker
	metadata metadata.Metadata
	logger   logger.Logger
}

func NewChain(name string, opts ...ChainOption) *Chain {
	var options ChainOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return &Chain{
		name:     name,
		metadata: options.Metadata,
		marker:   selector.NewFailMarker(),
		logger:   options.Logger,
	}
}

func (c *Chain) AddHop(hop chain.Hop) {
	c.hops = append(c.hops, hop)
}

// Metadata implements metadata.Metadatable interface.
func (c *Chain) Metadata() metadata.Metadata {
	return c.metadata
}

// Marker implements selector.Markable interface.
func (c *Chain) Marker() selector.Marker {
	return c.marker
}

func (c *Chain) Name() string {
	return c.name
}

func (c *Chain) Route(ctx context.Context, network, address string) chain.Route {
	if c == nil || len(c.hops) == 0 {
		return nil
	}

	rt := NewRoute(ChainRouteOption(c))
	for _, hop := range c.hops {
		node := hop.Select(ctx, chain.AddrSelectOption(address))
		if node == nil {
			return rt
		}
		if node.Options().Transport.Multiplex() {
			tr := node.Options().Transport.Copy()
			tr.Options().Route = rt
			node = node.Copy()
			node.Options().Transport = tr
			rt = NewRoute()
		}

		rt.addNode(node)
	}
	return rt
}

type chainGroup struct {
	chains   []chain.Chainer
	selector selector.Selector[chain.Chainer]
}

func NewChainGroup(chains ...chain.Chainer) *chainGroup {
	return &chainGroup{chains: chains}
}

func (p *chainGroup) WithSelector(s selector.Selector[chain.Chainer]) *chainGroup {
	p.selector = s
	return p
}

func (p *chainGroup) Route(ctx context.Context, network, address string) chain.Route {
	if chain := p.next(ctx); chain != nil {
		return chain.Route(ctx, network, address)
	}
	return nil
}

func (p *chainGroup) next(ctx context.Context) chain.Chainer {
	if p == nil || len(p.chains) == 0 {
		return nil
	}

	return p.selector.Select(ctx, p.chains...)
}

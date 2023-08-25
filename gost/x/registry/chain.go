package registry

import (
	"context"

	"proxy_forwarder/gost/core/chain"
	"proxy_forwarder/gost/core/metadata"
	"proxy_forwarder/gost/core/selector"
)

type chainRegistry struct {
	registry[chain.Chainer]
}

func (r *chainRegistry) Register(name string, v chain.Chainer) error {
	return r.registry.Register(name, v)
}

func (r *chainRegistry) Get(name string) chain.Chainer {
	if name != "" {
		return &chainWrapper{name: name, r: r}
	}
	return nil
}

func (r *chainRegistry) get(name string) chain.Chainer {
	return r.registry.Get(name)
}

type chainWrapper struct {
	name string
	r    *chainRegistry
}

func (w *chainWrapper) Marker() selector.Marker {
	v := w.r.get(w.name)
	if v == nil {
		return nil
	}
	if mi, ok := v.(selector.Markable); ok {
		return mi.Marker()
	}
	return nil
}

func (w *chainWrapper) Metadata() metadata.Metadata {
	v := w.r.get(w.name)
	if v == nil {
		return nil
	}

	if mi, ok := v.(metadata.Metadatable); ok {
		return mi.Metadata()
	}
	return nil
}

func (w *chainWrapper) Route(ctx context.Context, network, address string) chain.Route {
	v := w.r.get(w.name)
	if v == nil {
		return nil
	}
	return v.Route(ctx, network, address)
}

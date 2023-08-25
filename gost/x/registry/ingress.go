package registry

import (
	"context"

	"proxy_forwarder/gost/core/ingress"
)

type ingressRegistry struct {
	registry[ingress.Ingress]
}

func (r *ingressRegistry) Register(name string, v ingress.Ingress) error {
	return r.registry.Register(name, v)
}

func (r *ingressRegistry) Get(name string) ingress.Ingress {
	if name != "" {
		return &ingressWrapper{name: name, r: r}
	}
	return nil
}

func (r *ingressRegistry) get(name string) ingress.Ingress {
	return r.registry.Get(name)
}

type ingressWrapper struct {
	name string
	r    *ingressRegistry
}

func (w *ingressWrapper) Get(ctx context.Context, host string) string {
	v := w.r.get(w.name)
	if v == nil {
		return ""
	}
	return v.Get(ctx, host)
}

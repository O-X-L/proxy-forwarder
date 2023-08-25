package registry

import (
	"proxy_forwarder/gost/core/limiter/conn"
	"proxy_forwarder/gost/core/limiter/rate"
	"proxy_forwarder/gost/core/limiter/traffic"
)

type trafficLimiterRegistry struct {
	registry[traffic.TrafficLimiter]
}

func (r *trafficLimiterRegistry) Register(name string, v traffic.TrafficLimiter) error {
	return r.registry.Register(name, v)
}

func (r *trafficLimiterRegistry) Get(name string) traffic.TrafficLimiter {
	if name != "" {
		return &trafficLimiterWrapper{name: name, r: r}
	}
	return nil
}

func (r *trafficLimiterRegistry) get(name string) traffic.TrafficLimiter {
	return r.registry.Get(name)
}

type trafficLimiterWrapper struct {
	name string
	r    *trafficLimiterRegistry
}

func (w *trafficLimiterWrapper) In(key string) traffic.Limiter {
	v := w.r.get(w.name)
	if v == nil {
		return nil
	}
	return v.In(key)
}

func (w *trafficLimiterWrapper) Out(key string) traffic.Limiter {
	v := w.r.get(w.name)
	if v == nil {
		return nil
	}
	return v.Out(key)
}

type connLimiterRegistry struct {
	registry[conn.ConnLimiter]
}

func (r *connLimiterRegistry) Register(name string, v conn.ConnLimiter) error {
	return r.registry.Register(name, v)
}

func (r *connLimiterRegistry) Get(name string) conn.ConnLimiter {
	if name != "" {
		return &connLimiterWrapper{name: name, r: r}
	}
	return nil
}

func (r *connLimiterRegistry) get(name string) conn.ConnLimiter {
	return r.registry.Get(name)
}

type connLimiterWrapper struct {
	name string
	r    *connLimiterRegistry
}

func (w *connLimiterWrapper) Limiter(key string) conn.Limiter {
	v := w.r.get(w.name)
	if v == nil {
		return nil
	}
	return v.Limiter(key)
}

type rateLimiterRegistry struct {
	registry[rate.RateLimiter]
}

func (r *rateLimiterRegistry) Register(name string, v rate.RateLimiter) error {
	return r.registry.Register(name, v)
}

func (r *rateLimiterRegistry) Get(name string) rate.RateLimiter {
	if name != "" {
		return &rateLimiterWrapper{name: name, r: r}
	}
	return nil
}

func (r *rateLimiterRegistry) get(name string) rate.RateLimiter {
	return r.registry.Get(name)
}

type rateLimiterWrapper struct {
	name string
	r    *rateLimiterRegistry
}

func (w *rateLimiterWrapper) Limiter(key string) rate.Limiter {
	v := w.r.get(w.name)
	if v == nil {
		return nil
	}
	return v.Limiter(key)
}

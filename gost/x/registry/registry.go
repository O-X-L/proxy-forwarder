package registry

import (
	"errors"
	"io"
	"sync"

	"proxy_forwarder/gost/core/admission"
	"proxy_forwarder/gost/core/auth"
	"proxy_forwarder/gost/core/bypass"
	"proxy_forwarder/gost/core/chain"
	"proxy_forwarder/gost/core/hosts"
	"proxy_forwarder/gost/core/ingress"
	"proxy_forwarder/gost/core/limiter/conn"
	"proxy_forwarder/gost/core/limiter/rate"
	"proxy_forwarder/gost/core/limiter/traffic"
	"proxy_forwarder/gost/core/recorder"
	reg "proxy_forwarder/gost/core/registry"
	"proxy_forwarder/gost/core/resolver"
	"proxy_forwarder/gost/core/service"
)

var (
	ErrDup = errors.New("registry: duplicate object")
)

var (
	listenerReg  reg.Registry[NewListener]         = new(listenerRegistry)
	handlerReg   reg.Registry[NewHandler]          = new(handlerRegistry)
	dialerReg    reg.Registry[NewDialer]           = new(dialerRegistry)
	connectorReg reg.Registry[NewConnector]        = new(connectorRegistry)
	serviceReg   reg.Registry[service.Service]     = new(serviceRegistry)
	chainReg     reg.Registry[chain.Chainer]       = new(chainRegistry)
	hopReg       reg.Registry[chain.Hop]           = new(hopRegistry)
	autherReg    reg.Registry[auth.Authenticator]  = new(autherRegistry)
	admissionReg reg.Registry[admission.Admission] = new(admissionRegistry)
	bypassReg    reg.Registry[bypass.Bypass]       = new(bypassRegistry)
	resolverReg  reg.Registry[resolver.Resolver]   = new(resolverRegistry)
	hostsReg     reg.Registry[hosts.HostMapper]    = new(hostsRegistry)
	recorderReg  reg.Registry[recorder.Recorder]   = new(recorderRegistry)

	trafficLimiterReg reg.Registry[traffic.TrafficLimiter] = new(trafficLimiterRegistry)
	connLimiterReg    reg.Registry[conn.ConnLimiter]       = new(connLimiterRegistry)
	rateLimiterReg    reg.Registry[rate.RateLimiter]       = new(rateLimiterRegistry)

	ingressReg reg.Registry[ingress.Ingress] = new(ingressRegistry)
)

type registry[T any] struct {
	m sync.Map
}

func (r *registry[T]) Register(name string, v T) error {
	if name == "" {
		return nil
	}
	if _, loaded := r.m.LoadOrStore(name, v); loaded {
		return ErrDup
	}

	return nil
}

func (r *registry[T]) Unregister(name string) {
	if v, ok := r.m.Load(name); ok {
		if closer, ok := v.(io.Closer); ok {
			closer.Close()
		}
		r.m.Delete(name)
	}
}

func (r *registry[T]) IsRegistered(name string) bool {
	_, ok := r.m.Load(name)
	return ok
}

func (r *registry[T]) Get(name string) (t T) {
	if name == "" {
		return
	}
	v, _ := r.m.Load(name)
	t, _ = v.(T)
	return
}

func (r *registry[T]) GetAll() (m map[string]T) {
	m = make(map[string]T)
	r.m.Range(func(key, value any) bool {
		k, _ := key.(string)
		v, _ := value.(T)
		m[k] = v
		return true
	})
	return
}

func ListenerRegistry() reg.Registry[NewListener] {
	return listenerReg
}

func HandlerRegistry() reg.Registry[NewHandler] {
	return handlerReg
}

func DialerRegistry() reg.Registry[NewDialer] {
	return dialerReg
}

func ConnectorRegistry() reg.Registry[NewConnector] {
	return connectorReg
}

func ServiceRegistry() reg.Registry[service.Service] {
	return serviceReg
}

func ChainRegistry() reg.Registry[chain.Chainer] {
	return chainReg
}

func HopRegistry() reg.Registry[chain.Hop] {
	return hopReg
}

func AutherRegistry() reg.Registry[auth.Authenticator] {
	return autherReg
}

func AdmissionRegistry() reg.Registry[admission.Admission] {
	return admissionReg
}

func BypassRegistry() reg.Registry[bypass.Bypass] {
	return bypassReg
}

func ResolverRegistry() reg.Registry[resolver.Resolver] {
	return resolverReg
}

func HostsRegistry() reg.Registry[hosts.HostMapper] {
	return hostsReg
}

func RecorderRegistry() reg.Registry[recorder.Recorder] {
	return recorderReg
}

func TrafficLimiterRegistry() reg.Registry[traffic.TrafficLimiter] {
	return trafficLimiterReg
}

func ConnLimiterRegistry() reg.Registry[conn.ConnLimiter] {
	return connLimiterReg
}

func RateLimiterRegistry() reg.Registry[rate.RateLimiter] {
	return rateLimiterReg
}

func IngressRegistry() reg.Registry[ingress.Ingress] {
	return ingressReg
}

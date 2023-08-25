package service

import (
	"net"
	"net/http"

	"proxy_forwarder/gost/core/service"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	DefaultPath = "/metrics"
)

type options struct {
	path string
}

type Option func(*options)

func PathOption(path string) Option {
	return func(o *options) {
		o.path = path
	}
}

type metricService struct {
	s  *http.Server
	ln net.Listener
}

func NewService(addr string, opts ...Option) (service.Service, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	var options options
	for _, opt := range opts {
		opt(&options)
	}
	if options.path == "" {
		options.path = DefaultPath
	}

	mux := http.NewServeMux()
	mux.Handle(options.path, promhttp.Handler())
	return &metricService{
		s: &http.Server{
			Handler: mux,
		},
		ln: ln,
	}, nil
}

func (s *metricService) Serve() error {
	return s.s.Serve(s.ln)
}

func (s *metricService) Addr() net.Addr {
	return s.ln.Addr()
}

func (s *metricService) Close() error {
	return s.s.Close()
}

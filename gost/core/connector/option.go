package connector

import (
	"crypto/tls"
	"net/url"
	"time"

	"proxy_forwarder/gost/core/common/net/dialer"
	"proxy_forwarder/gost/core/logger"
)

type Options struct {
	Auth      *url.Userinfo
	TLSConfig *tls.Config
	Logger    logger.Logger
}

type Option func(opts *Options)

func AuthOption(auth *url.Userinfo) Option {
	return func(opts *Options) {
		opts.Auth = auth
	}
}

func TLSConfigOption(tlsConfig *tls.Config) Option {
	return func(opts *Options) {
		opts.TLSConfig = tlsConfig
	}
}

func LoggerOption(logger logger.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}

type ConnectOptions struct {
	NetDialer *dialer.NetDialer
}

type ConnectOption func(opts *ConnectOptions)

func NetDialerConnectOption(netd *dialer.NetDialer) ConnectOption {
	return func(opts *ConnectOptions) {
		opts.NetDialer = netd
	}
}

type BindOptions struct {
	Mux               bool
	Backlog           int
	UDPDataQueueSize  int
	UDPDataBufferSize int
	UDPConnTTL        time.Duration
}

type BindOption func(opts *BindOptions)

func MuxBindOption(mux bool) BindOption {
	return func(opts *BindOptions) {
		opts.Mux = mux
	}
}

func BacklogBindOption(backlog int) BindOption {
	return func(opts *BindOptions) {
		opts.Backlog = backlog
	}
}

func UDPDataQueueSizeBindOption(size int) BindOption {
	return func(opts *BindOptions) {
		opts.UDPDataQueueSize = size
	}
}

func UDPDataBufferSizeBindOption(size int) BindOption {
	return func(opts *BindOptions) {
		opts.UDPDataBufferSize = size
	}
}

func UDPConnTTLBindOption(ttl time.Duration) BindOption {
	return func(opts *BindOptions) {
		opts.UDPConnTTL = ttl
	}
}

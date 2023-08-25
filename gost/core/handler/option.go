package handler

import (
	"crypto/tls"
	"net/url"

	"proxy_forwarder/gost/core/auth"
	"proxy_forwarder/gost/core/bypass"
	"proxy_forwarder/gost/core/chain"
	"proxy_forwarder/gost/core/limiter/rate"
	"proxy_forwarder/gost/core/logger"
	"proxy_forwarder/gost/core/metadata"
)

type Options struct {
	Bypass      bypass.Bypass
	Router      *chain.Router
	Auth        *url.Userinfo
	Auther      auth.Authenticator
	RateLimiter rate.RateLimiter
	TLSConfig   *tls.Config
	Logger      logger.Logger
	Service     string
}

type Option func(opts *Options)

func BypassOption(bypass bypass.Bypass) Option {
	return func(opts *Options) {
		opts.Bypass = bypass
	}
}

func RouterOption(router *chain.Router) Option {
	return func(opts *Options) {
		opts.Router = router
	}
}

func AuthOption(auth *url.Userinfo) Option {
	return func(opts *Options) {
		opts.Auth = auth
	}
}

func AutherOption(auther auth.Authenticator) Option {
	return func(opts *Options) {
		opts.Auther = auther
	}
}

func RateLimiterOption(limiter rate.RateLimiter) Option {
	return func(opts *Options) {
		opts.RateLimiter = limiter
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

func ServiceOption(service string) Option {
	return func(opts *Options) {
		opts.Service = service
	}
}

type HandleOptions struct {
	Metadata metadata.Metadata
}

type HandleOption func(opts *HandleOptions)

func MetadataHandleOption(md metadata.Metadata) HandleOption {
	return func(opts *HandleOptions) {
		opts.Metadata = md
	}
}

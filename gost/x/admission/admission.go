package admission

import (
	"bufio"
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	admission_pkg "proxy_forwarder/gost/core/admission"
	"proxy_forwarder/gost/core/logger"
	"proxy_forwarder/gost/x/internal/loader"
	"proxy_forwarder/gost/x/internal/matcher"

	"google.golang.org/grpc"
)

type options struct {
	whitelist   bool
	matchers    []string
	fileLoader  loader.Loader
	redisLoader loader.Loader
	httpLoader  loader.Loader
	client      *grpc.ClientConn
	period      time.Duration
	logger      logger.Logger
}

type Option func(opts *options)

func WhitelistOption(whitelist bool) Option {
	return func(opts *options) {
		opts.whitelist = whitelist
	}
}

func MatchersOption(matchers []string) Option {
	return func(opts *options) {
		opts.matchers = matchers
	}
}

func ReloadPeriodOption(period time.Duration) Option {
	return func(opts *options) {
		opts.period = period
	}
}

func FileLoaderOption(fileLoader loader.Loader) Option {
	return func(opts *options) {
		opts.fileLoader = fileLoader
	}
}

func RedisLoaderOption(redisLoader loader.Loader) Option {
	return func(opts *options) {
		opts.redisLoader = redisLoader
	}
}

func HTTPLoaderOption(httpLoader loader.Loader) Option {
	return func(opts *options) {
		opts.httpLoader = httpLoader
	}
}

func PluginConnOption(c *grpc.ClientConn) Option {
	return func(opts *options) {
		opts.client = c
	}
}

func LoggerOption(logger logger.Logger) Option {
	return func(opts *options) {
		opts.logger = logger
	}
}

type admission struct {
	ipMatcher   matcher.Matcher
	cidrMatcher matcher.Matcher
	mu          sync.RWMutex
	cancelFunc  context.CancelFunc
	options     options
}

// NewAdmission creates and initializes a new Admission using matcher patterns as its match rules.
// The rules will be reversed if the reverse is true.
func NewAdmission(opts ...Option) admission_pkg.Admission {
	var options options
	for _, opt := range opts {
		opt(&options)
	}

	ctx, cancel := context.WithCancel(context.TODO())
	p := &admission{
		cancelFunc: cancel,
		options:    options,
	}

	if err := p.reload(ctx); err != nil {
		options.logger.Warnf("reload: %v", err)
	}
	if p.options.period > 0 {
		go p.periodReload(ctx)
	}

	return p
}

func (p *admission) Admit(ctx context.Context, addr string) bool {
	if addr == "" || p == nil {
		return true
	}

	// try to strip the port
	if host, _, _ := net.SplitHostPort(addr); host != "" {
		addr = host
	}

	matched := p.matched(addr)

	return !p.options.whitelist && !matched ||
		p.options.whitelist && matched
}

func (p *admission) periodReload(ctx context.Context) error {
	period := p.options.period
	if period < time.Second {
		period = time.Second
	}
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.reload(ctx); err != nil {
				p.options.logger.Warnf("reload: %v", err)
				// return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (p *admission) reload(ctx context.Context) error {
	v, err := p.load(ctx)
	if err != nil {
		return err
	}
	patterns := append(p.options.matchers, v...)

	var ips []net.IP
	var inets []*net.IPNet
	for _, pattern := range patterns {
		if ip := net.ParseIP(pattern); ip != nil {
			ips = append(ips, ip)
			continue
		}
		if _, inet, err := net.ParseCIDR(pattern); err == nil {
			inets = append(inets, inet)
			continue
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.ipMatcher = matcher.IPMatcher(ips)
	p.cidrMatcher = matcher.CIDRMatcher(inets)

	return nil
}

func (p *admission) load(ctx context.Context) (patterns []string, err error) {
	if p.options.fileLoader != nil {
		if lister, ok := p.options.fileLoader.(loader.Lister); ok {
			list, er := lister.List(ctx)
			if er != nil {
				p.options.logger.Warnf("file loader: %v", er)
			}
			for _, s := range list {
				if line := p.parseLine(s); line != "" {
					patterns = append(patterns, line)
				}
			}
		} else {
			r, er := p.options.fileLoader.Load(ctx)
			if er != nil {
				p.options.logger.Warnf("file loader: %v", er)
			}
			if v, _ := p.parsePatterns(r); v != nil {
				patterns = append(patterns, v...)
			}
		}
	}
	if p.options.redisLoader != nil {
		if lister, ok := p.options.redisLoader.(loader.Lister); ok {
			list, er := lister.List(ctx)
			if er != nil {
				p.options.logger.Warnf("redis loader: %v", er)
			}
			patterns = append(patterns, list...)
		} else {
			r, er := p.options.redisLoader.Load(ctx)
			if er != nil {
				p.options.logger.Warnf("redis loader: %v", er)
			}
			if v, _ := p.parsePatterns(r); v != nil {
				patterns = append(patterns, v...)
			}
		}
	}

	if p.options.httpLoader != nil {
		r, er := p.options.httpLoader.Load(ctx)
		if er != nil {
			p.options.logger.Warnf("http loader: %v", er)
		}
		if v, _ := p.parsePatterns(r); v != nil {
			patterns = append(patterns, v...)
		}
	}

	p.options.logger.Debugf("load items %d", len(patterns))
	return
}

func (p *admission) parsePatterns(r io.Reader) (patterns []string, err error) {
	if r == nil {
		return
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if line := p.parseLine(scanner.Text()); line != "" {
			patterns = append(patterns, line)
		}
	}

	err = scanner.Err()
	return
}

func (p *admission) parseLine(s string) string {
	if n := strings.IndexByte(s, '#'); n >= 0 {
		s = s[:n]
	}
	return strings.TrimSpace(s)
}

func (p *admission) matched(addr string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.ipMatcher.Match(addr) ||
		p.cidrMatcher.Match(addr)
}

func (p *admission) Close() error {
	p.cancelFunc()
	if p.options.fileLoader != nil {
		p.options.fileLoader.Close()
	}
	if p.options.redisLoader != nil {
		p.options.redisLoader.Close()
	}
	return nil
}

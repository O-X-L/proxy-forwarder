package udp

import (
	"sync"
	"time"

	"proxy_forwarder/gost/core/logger"
)

type connPool struct {
	m      sync.Map
	ttl    time.Duration
	closed chan struct{}
	logger logger.Logger
}

func newConnPool(ttl time.Duration) *connPool {
	p := &connPool{
		ttl:    ttl,
		closed: make(chan struct{}),
	}
	go p.idleCheck()
	return p
}

func (p *connPool) WithLogger(logger logger.Logger) *connPool {
	p.logger = logger
	return p
}

func (p *connPool) Get(key any) (c *conn, ok bool) {
	v, ok := p.m.Load(key)
	if ok {
		c, ok = v.(*conn)
	}
	return
}

func (p *connPool) Set(key any, c *conn) {
	p.m.Store(key, c)
}

func (p *connPool) Delete(key any) {
	p.m.Delete(key)
}

func (p *connPool) Close() {
	select {
	case <-p.closed:
		return
	default:
	}

	close(p.closed)

	p.m.Range(func(k, v any) bool {
		if c, ok := v.(*conn); ok && c != nil {
			c.Close()
		}
		return true
	})
}

func (p *connPool) idleCheck() {
	ticker := time.NewTicker(p.ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			size := 0
			idles := 0
			p.m.Range(func(key, value any) bool {
				c, ok := value.(*conn)
				if !ok || c == nil {
					p.Delete(key)
					return true
				}
				size++

				if c.IsIdle() {
					idles++
					p.Delete(key)
					c.Close()
					return true
				}

				c.SetIdle(true)

				return true
			})

			if idles > 0 {
				p.logger.Debugf("connection pool: size=%d, idle=%d", size, idles)
			}
		case <-p.closed:
			return
		}
	}
}

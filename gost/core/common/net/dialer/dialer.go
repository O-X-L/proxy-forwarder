package dialer

import (
	"context"
	"fmt"
	"net"
	"strings"
	"syscall"
	"time"

	xnet "proxy_forwarder/gost/core/common/net"
	"proxy_forwarder/log"
)

const (
	DefaultTimeout = 10 * time.Second
)

var (
	DefaultNetDialer = &NetDialer{}
)

type NetDialer struct {
	Interface string
	Mark      int
	Timeout   time.Duration
	DialFunc  func(ctx context.Context, network, addr string) (net.Conn, error)
}

func (d *NetDialer) Dial(ctx context.Context, network, addr string) (conn net.Conn, err error) {
	if d == nil {
		d = DefaultNetDialer
	}

	timeout := d.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	if d.DialFunc != nil {
		return d.DialFunc(ctx, network, addr)
	}

	deadline := time.Now().Add(timeout)
	ifces := strings.Split(d.Interface, ",")
	for _, ifce := range ifces {
		strict := strings.HasSuffix(ifce, "!")
		ifce = strings.TrimSuffix(ifce, "!")
		var ifceName string
		var ifAddrs []net.Addr
		ifceName, ifAddrs, err = xnet.ParseInterfaceAddr(ifce, network)
		if err != nil && strict {
			return
		}

		for _, ifAddr := range ifAddrs {
			conn, err = d.dialOnce(ctx, network, addr, ifceName, ifAddr, deadline)
			if err == nil {
				return
			}

			log.Debug("dialer", fmt.Sprintf("dial %s %v@%s failed: %s", network, ifAddr, ifceName, err))

			if strict &&
				!strings.Contains(err.Error(), "no suitable address found") &&
				!strings.Contains(err.Error(), "mismatched local address type") {
				return
			}

			if time.Until(deadline) < 0 {
				return
			}
		}
	}

	return
}

func (d *NetDialer) dialOnce(ctx context.Context, network, addr, ifceName string, ifAddr net.Addr, deadline time.Time) (net.Conn, error) {
	if ifceName != "" {
		log.Debug("dialer", fmt.Sprintf("interface: %s %v/%s", ifceName, ifAddr, network))
	}

	switch network {
	case "udp", "udp4", "udp6":
		if addr == "" {
			var laddr *net.UDPAddr
			if ifAddr != nil {
				laddr, _ = ifAddr.(*net.UDPAddr)
			}

			c, err := net.ListenUDP(network, laddr)
			if err != nil {
				return nil, err
			}
			sc, err := c.SyscallConn()
			if err != nil {
				log.Error("dialer", err)
				return nil, err
			}
			err = sc.Control(func(fd uintptr) {
				if ifceName != "" {
					if err := bindDevice(fd, ifceName); err != nil {
						log.Info("dialer", fmt.Sprintf("bind device: %v", err))
					}
				}
				if d.Mark != 0 {
					if err := setMark(fd, d.Mark); err != nil {
						log.Info("dialer", fmt.Sprintf("set mark: %v", err))
					}
				}
			})
			if err != nil {
				log.Error("dialer", err)
			}
			return c, nil
		}
	case "tcp", "tcp4", "tcp6":
	default:
		return nil, fmt.Errorf("dial: unsupported network %s", network)
	}
	netd := net.Dialer{
		Deadline:  deadline,
		LocalAddr: ifAddr,
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				if ifceName != "" {
					if err := bindDevice(fd, ifceName); err != nil {
						log.Info("dialer", fmt.Sprintf("bind device: %v", err))
					}
				}
				if d.Mark != 0 {
					if err := setMark(fd, d.Mark); err != nil {
						log.Info("dialer", fmt.Sprintf("set mark: %v", err))
					}
				}
			})
		},
	}
	return netd.DialContext(ctx, network, addr)
}

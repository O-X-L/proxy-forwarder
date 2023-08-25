package wrapper

import (
	"errors"
	"io"
	"net"
	"syscall"

	"proxy_forwarder/gost/core/metadata"
	"proxy_forwarder/gost/core/metrics"
	xnet "proxy_forwarder/gost/x/internal/net"
	"proxy_forwarder/gost/x/internal/net/udp"
	xmetrics "proxy_forwarder/gost/x/metrics"
)

var (
	errUnsupport = errors.New("unsupported operation")
)

// serverConn is a server side Conn with metrics supported.
type serverConn struct {
	net.Conn
	service string
}

func WrapConn(service string, c net.Conn) net.Conn {
	if !xmetrics.IsEnabled() {
		return c
	}
	return &serverConn{
		service: service,
		Conn:    c,
	}
}

func (c *serverConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if counter := xmetrics.GetCounter(
		xmetrics.MetricServiceTransferInputBytesCounter,
		metrics.Labels{
			"service": c.service,
		}); counter != nil {
		counter.Add(float64(n))
	}
	return
}

func (c *serverConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if counter := xmetrics.GetCounter(
		xmetrics.MetricServiceTransferOutputBytesCounter,
		metrics.Labels{
			"service": c.service,
		}); counter != nil {
		counter.Add(float64(n))
	}
	return
}

func (c *serverConn) SyscallConn() (rc syscall.RawConn, err error) {
	if sc, ok := c.Conn.(syscall.Conn); ok {
		rc, err = sc.SyscallConn()
		return
	}
	err = errUnsupport
	return
}

func (c *serverConn) Metadata() metadata.Metadata {
	if md, ok := c.Conn.(metadata.Metadatable); ok {
		return md.Metadata()
	}
	return nil
}

type packetConn struct {
	net.PacketConn
	service string
}

func WrapPacketConn(service string, pc net.PacketConn) net.PacketConn {
	if !xmetrics.IsEnabled() {
		return pc
	}
	return &packetConn{
		PacketConn: pc,
		service:    service,
	}
}

func (c *packetConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, addr, err = c.PacketConn.ReadFrom(p)
	if counter := xmetrics.GetCounter(
		xmetrics.MetricServiceTransferInputBytesCounter,
		metrics.Labels{
			"service": c.service,
		}); counter != nil {
		counter.Add(float64(n))
	}
	return
}

func (c *packetConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	n, err = c.PacketConn.WriteTo(p, addr)
	if counter := xmetrics.GetCounter(
		xmetrics.MetricServiceTransferOutputBytesCounter,
		metrics.Labels{
			"service": c.service,
		}); counter != nil {
		counter.Add(float64(n))
	}
	return
}

func (c *packetConn) Metadata() metadata.Metadata {
	if md, ok := c.PacketConn.(metadata.Metadatable); ok {
		return md.Metadata()
	}
	return nil
}

type udpConn struct {
	net.PacketConn
	service string
}

func WrapUDPConn(service string, pc net.PacketConn) udp.Conn {
	return &udpConn{
		PacketConn: pc,
		service:    service,
	}
}

func (c *udpConn) RemoteAddr() net.Addr {
	if nc, ok := c.PacketConn.(xnet.RemoteAddr); ok {
		return nc.RemoteAddr()
	}
	return nil
}

func (c *udpConn) SetReadBuffer(n int) error {
	if nc, ok := c.PacketConn.(xnet.SetBuffer); ok {
		return nc.SetReadBuffer(n)
	}
	return errUnsupport
}

func (c *udpConn) SetWriteBuffer(n int) error {
	if nc, ok := c.PacketConn.(xnet.SetBuffer); ok {
		return nc.SetWriteBuffer(n)
	}
	return errUnsupport
}

func (c *udpConn) Read(b []byte) (n int, err error) {
	if nc, ok := c.PacketConn.(io.Reader); ok {
		n, err = nc.Read(b)
		if counter := xmetrics.GetCounter(
			xmetrics.MetricServiceTransferInputBytesCounter,
			metrics.Labels{
				"service": c.service,
			}); counter != nil {
			counter.Add(float64(n))
		}
		return
	}
	err = errUnsupport
	return
}

func (c *udpConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, addr, err = c.PacketConn.ReadFrom(p)
	if counter := xmetrics.GetCounter(
		xmetrics.MetricServiceTransferInputBytesCounter,
		metrics.Labels{
			"service": c.service,
		}); counter != nil {
		counter.Add(float64(n))
	}
	return
}

func (c *udpConn) ReadFromUDP(b []byte) (n int, addr *net.UDPAddr, err error) {
	if nc, ok := c.PacketConn.(udp.ReadUDP); ok {
		n, addr, err = nc.ReadFromUDP(b)
		if counter := xmetrics.GetCounter(
			xmetrics.MetricServiceTransferInputBytesCounter,
			metrics.Labels{
				"service": c.service,
			}); counter != nil {
			counter.Add(float64(n))
		}
		return
	}
	err = errUnsupport
	return
}

func (c *udpConn) ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *net.UDPAddr, err error) {
	if nc, ok := c.PacketConn.(udp.ReadUDP); ok {
		n, oobn, flags, addr, err = nc.ReadMsgUDP(b, oob)
		if counter := xmetrics.GetCounter(
			xmetrics.MetricServiceTransferInputBytesCounter,
			metrics.Labels{
				"service": c.service,
			}); counter != nil {
			counter.Add(float64(n))
		}
		return
	}
	err = errUnsupport
	return
}

func (c *udpConn) Write(b []byte) (n int, err error) {
	if nc, ok := c.PacketConn.(io.Writer); ok {
		n, err = nc.Write(b)
		if counter := xmetrics.GetCounter(
			xmetrics.MetricServiceTransferOutputBytesCounter,
			metrics.Labels{
				"service": c.service,
			}); counter != nil {
			counter.Add(float64(n))
		}
		return
	}
	err = errUnsupport
	return
}

func (c *udpConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	n, err = c.PacketConn.WriteTo(p, addr)
	if counter := xmetrics.GetCounter(
		xmetrics.MetricServiceTransferOutputBytesCounter,
		metrics.Labels{
			"service": c.service,
		}); counter != nil {
		counter.Add(float64(n))
	}
	return
}

func (c *udpConn) WriteToUDP(b []byte, addr *net.UDPAddr) (n int, err error) {
	if nc, ok := c.PacketConn.(udp.WriteUDP); ok {
		n, err = nc.WriteToUDP(b, addr)
		if counter := xmetrics.GetCounter(
			xmetrics.MetricServiceTransferOutputBytesCounter,
			metrics.Labels{
				"service": c.service,
			}); counter != nil {
			counter.Add(float64(n))
		}
		return
	}
	err = errUnsupport
	return
}

func (c *udpConn) WriteMsgUDP(b, oob []byte, addr *net.UDPAddr) (n, oobn int, err error) {
	if nc, ok := c.PacketConn.(udp.WriteUDP); ok {
		n, oobn, err = nc.WriteMsgUDP(b, oob, addr)
		if counter := xmetrics.GetCounter(
			xmetrics.MetricServiceTransferOutputBytesCounter,
			metrics.Labels{
				"service": c.service,
			}); counter != nil {
			counter.Add(float64(n))
		}
		return
	}
	err = errUnsupport
	return
}

func (c *udpConn) SyscallConn() (rc syscall.RawConn, err error) {
	if nc, ok := c.PacketConn.(syscall.Conn); ok {
		return nc.SyscallConn()
	}
	err = errUnsupport
	return
}

func (c *udpConn) SetDSCP(n int) error {
	if nc, ok := c.PacketConn.(xnet.SetDSCP); ok {
		return nc.SetDSCP(n)
	}
	return nil
}

func (c *udpConn) Metadata() metadata.Metadata {
	if md, ok := c.PacketConn.(metadata.Metadatable); ok {
		return md.Metadata()
	}
	return nil
}

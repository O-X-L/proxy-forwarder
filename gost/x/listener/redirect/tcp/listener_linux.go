package tcp

import (
	"fmt"
	"proxy_forwarder/log"
	"syscall"

	"golang.org/x/sys/unix"
)

func (l *redirectListener) control(network, address string, c syscall.RawConn) error {
	return c.Control(func(fd uintptr) {
		if err := unix.SetsockoptInt(int(fd), unix.SOL_IP, unix.IP_TRANSPARENT, 1); err != nil {
			log.ErrorS("listener", fmt.Sprintf("SetsockoptInt(SOL_IP, IP_TRANSPARENT, 1): %v", err))
		}
	})
}

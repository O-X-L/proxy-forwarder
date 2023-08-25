package main

import (
	// Register connectors
	_ "proxy_forwarder/gost/x/connector/http"

	// Register dialers
	_ "proxy_forwarder/gost/x/dialer/direct"
	_ "proxy_forwarder/gost/x/dialer/tcp"
	_ "proxy_forwarder/gost/x/dialer/tls"
	_ "proxy_forwarder/gost/x/dialer/udp"

	// Register handlers
	_ "proxy_forwarder/gost/x/handler/auto"
	_ "proxy_forwarder/gost/x/handler/http"
	_ "proxy_forwarder/gost/x/handler/redirect/tcp"
	_ "proxy_forwarder/gost/x/handler/redirect/udp"
	_ "proxy_forwarder/gost/x/handler/sni"

	// Register listeners
	_ "proxy_forwarder/gost/x/listener/redirect/tcp"
	_ "proxy_forwarder/gost/x/listener/redirect/udp"
)

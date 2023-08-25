package registry

import (
	"proxy_forwarder/gost/core/service"
)

type serviceRegistry struct {
	registry[service.Service]
}

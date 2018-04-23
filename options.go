package servers

import (
	"github.com/bweston92/healthz/healthz"
)

// WithInternalServerAddr sets the address to bind the internal
// use server too.
func WithInternalServerAddr(addr string) Option {
	return func(o *serverOptions) error {
		o.addr = addr
		return nil
	}
}

// WithHealthzComponent will add the healthz component to the healthz
// handler on the internal use server.
func WithHealthzComponent(c *healthz.Component) Option {
	return func(o *serverOptions) error {
		o.healthzComponents = append(o.healthzComponents, c)
		return nil
	}
}

// AddHealthzMetadata will add the metadata to the healthz handler
// on the internal use server.
func AddHealthzMetadata(k, v string) Option {
	return func(o *serverOptions) error {
		o.healthzMeta[k] = v
		return nil
	}
}

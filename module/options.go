package module

import (
	"time"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/selector"
	"github.com/nats-io/go-nats"
)


type Option func(*Options)

type Options struct {
	Nats 		*nats.Conn
	Version      	string
	Registry  	registry.Registry
	Selector  	selector.Selector
	// Register loop interval
	RegisterInterval time.Duration
	RegisterTTL time.Duration

}



func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

func Nats(nc *nats.Conn) Option {
	return func(o *Options) {
		o.Nats = nc
	}
}

// Registry sets the registry for the service
// and the underlying components
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
		o.Selector.Init(selector.Registry(r))
	}
}

func Selector(r selector.Selector) Option {
	return func(o *Options) {
		o.Selector = r
	}
}

// RegisterTTL specifies the TTL to use when registering the service
func RegisterTTL(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterTTL = t
	}
}

// RegisterInterval specifies the interval on which to re-register
func RegisterInterval(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterInterval = t
	}
}
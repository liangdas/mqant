package consul

import ()
import "github.com/liangdas/mqant/registry"

func NewRegistry(opts ...registry.Option) registry.Registry {
	return registry.NewRegistry(opts...)
}

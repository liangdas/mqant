package selector

import (
	"context"

	"github.com/liangdas/mqant/registry"
)

// Options Options
type Options struct {
	Watcher  Watcher
	Registry registry.Registry
	Strategy Strategy

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

// SelectOptions SelectOptions
type SelectOptions struct {
	Filters  []Filter
	Strategy Strategy

	Param []interface{}
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

// Watcher Watcher
type Watcher func(node *registry.Node)

// Option used to initialise the selector
type Option func(*Options)

// SelectOption used when making a select call
type SelectOption func(*SelectOptions)

// Registry sets the registry used by the selector
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

// SetStrategy sets the default strategy for the selector
func SetStrategy(fn Strategy) Option {
	return func(o *Options) {
		o.Strategy = fn
	}
}

// SetWatcher sets the default strategy for the selector
func SetWatcher(fn Watcher) Option {
	return func(o *Options) {
		o.Watcher = fn
	}
}

// WithFilter adds a filter function to the list of filters
// used during the Select call.
func WithFilter(fn ...Filter) SelectOption {
	return func(o *SelectOptions) {
		o.Filters = append(o.Filters, fn...)
	}
}

// WithStrategy sets the selector strategy
func WithStrategy(fn Strategy) SelectOption {
	return func(o *SelectOptions) {
		o.Strategy = fn
	}
}

package client

import (
	"context"

	"github.com/robert-pkg/base4go/registry"
)

// NewOptions creates new server options.
func NewOptions(opt ...Option) Options {
	opts := Options{
		Registry: registry.DefaultRegistry,
		Context:  context.Background(),
	}

	for _, o := range opt {
		o(&opts)
	}

	return opts
}

type Option func(*Options)

type Options struct {
	Registry registry.Registry

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

// Registry used for discovery.
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

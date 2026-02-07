package server

import (
	"context"
	"time"

	"github.com/robert-pkg/base4go/registry"
)

// NewOptions creates new server options.
func NewOptions(opt ...Option) Options {
	opts := Options{
		Registry:         registry.DefaultRegistry,
		Context:          context.Background(),
		RegisterInterval: time.Second * 30,
		RegisterTTL:      time.Second * 90,
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

	Host     string // 主机
	Port     int    // 端口
	HttpPort int    // 该http端口可用于metrics，服务健康检查

	// The interval on which to register
	RegisterInterval time.Duration

	// The register expiry time
	RegisterTTL time.Duration
}

// Registry used for discovery.
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

func Host(host string) Option {
	return func(o *Options) {
		o.Host = host
	}
}

func Port(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

func HttpPort(port int) Option {
	return func(o *Options) {
		o.HttpPort = port
	}
}

func RegisterInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.RegisterInterval = interval
	}
}

func RegisterTTL(interval time.Duration) Option {
	return func(o *Options) {
		o.RegisterTTL = interval
	}
}

type StartOptions struct {
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

type StartOption func(*StartOptions)

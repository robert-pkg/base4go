package app

import (
	"context"
	"os"

	"github.com/robert-pkg/base4go/registry"
	"github.com/robert-pkg/base4go/rpc/server"
)

type ServerInfo struct {
	Protocol string // grpc, http
	Host     string //

	StartOpts []server.StartOption
}

// Option is an application option.
type Option func(o *options)

// options is an application options.
type options struct {
	ctx  context.Context
	name string

	// log
	logFileName            string
	logEnableConsoleOutput bool
	logFields              map[string]interface{}

	registry registry.Registry

	sigs []os.Signal
}

// Context with application context.
func Context(ctx context.Context) Option {
	return func(o *options) { o.ctx = ctx }
}

// Name with application name.
func Name(name string) Option {
	return func(o *options) { o.name = name }
}

func LogFileName(v string) Option {
	return func(o *options) { o.logFileName = v }
}

func LogEnableConsoleOutput(v bool) Option {
	return func(o *options) { o.logEnableConsoleOutput = v }
}

func LogFields(v map[string]interface{}) Option {
	return func(o *options) { o.logFields = v }
}

// Signal with exit signals.
func Signal(sigs ...os.Signal) Option {
	return func(o *options) { o.sigs = sigs }
}

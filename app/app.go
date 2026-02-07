package app

import (
	"context"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/robert-pkg/base4go/log"
	"github.com/robert-pkg/base4go/registry"
	consul_registry "github.com/robert-pkg/base4go/registry/consul"
	"github.com/robert-pkg/base4go/rpc/server"
	"github.com/robert-pkg/base4go/rpc/server/grpc_server"
)

func New(opts ...Option) *App {
	o := options{
		ctx:                    context.Background(),
		logEnableConsoleOutput: true,
		logFields:              make(map[string]interface{}),
		sigs:                   []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
	}

	for _, opt := range opts {
		opt(&o)
	}

	ctx, cancel := context.WithCancel(o.ctx)
	return &App{
		ctx:    ctx,
		cancel: cancel,
		opts:   o,
	}
}

// App is an application components lifecycle manager.
type App struct {
	opts   options
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex
}

func (a *App) Name() string { return a.opts.name }

func (a *App) Run(serverInfo *ServerInfo) {

	err := initLog(a.opts.logFileName, a.opts.logEnableConsoleOutput, a.opts.logFields)
	if err != nil {
		panic(err)
	}

	if a.opts.registry == nil {
		registry.DefaultRegistry = consul_registry.NewRegistry(
			registry.Addrs("127.0.0.1:8500"),
		)
	}

	var svr server.Server
	if serverInfo.Protocol == "grpc" {
		gs := grpc_server.NewServer(
			server.Registry(registry.DefaultRegistry),
			server.Host(serverInfo.Host),
		)

		svr = gs
	} else {
		panic("no support")
	}

	err = svr.Start(serverInfo.StartOpts...)
	if err != nil {
		log.Errorf("Start fail: %v", err)
		os.Exit(1)
	}

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("process crash. err:%v, stack:%v", err, string(debug.Stack()))
		}

		svr.Stop()
		log.Infof("exit...")
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, a.opts.sigs...)

	select {
	case s := <-ch:
		log.Infof("recv quit signal, signal: %v", s)
	}

}

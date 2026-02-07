package app

import (
	"context"
	"fmt"
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

func (a *App) Init() error {
	err := InitLog(a.opts.logFileName, a.opts.logEnableConsoleOutput, a.opts.logFields)
	if err != nil {
		return err
	}

	if a.opts.registry == nil {
		registry.DefaultRegistry = consul_registry.NewRegistry(
			registry.Addrs("127.0.0.1:8500"),
		)
	}

	return nil
}

func (a *App) Run(serverInfo *ServerInfo) error {

	var svr server.Server
	if serverInfo.Protocol == "grpc" {
		gs := grpc_server.NewServer(
			server.Registry(registry.DefaultRegistry),
			server.Host(serverInfo.Host),
		)

		if err := gs.Init(); err != nil {
			return err
		}

		svr = gs
	} else {
		return fmt.Errorf("no support")
	}

	err := svr.Start(serverInfo.StartOpts...)
	if err != nil {
		log.Errorf("Start fail: %v", err)
		return err
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

	return nil
}

package app

import (
	"flag"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/robert-pkg/base4go/app"
	"github.com/robert-pkg/base4go/log"
	"github.com/robert-pkg/base4go/registry"
	consul_registry "github.com/robert-pkg/base4go/registry/consul"

	"github.com/robert-pkg/base4go/cmd/gateway/internal/config"
	"github.com/robert-pkg/base4go/cmd/gateway/internal/handler"
	"github.com/robert-pkg/base4go/cmd/gateway/internal/http_server"
)

func Run() {
	logFile := flag.String("log_file", "/var/log/gateway/gateway.log", "日志文件路径 (可选，默认 /var/log/gateway/gateway.log)")
	outputConsole := flag.Bool("output_console", false, "是否输出日志到环境变量 (可选，默认 false)")
	configPath := flag.String("config", "config.yaml", "配置文件路径 (可选，默认 config.yaml)")
	envPrefix := flag.String("env_prerix", "gateway", "环境变量前缀 (可选，默认 gateway)")

	flag.Parse()

	defer func() {
		if err := recover(); err != nil {
			log.Errorf("err. statck:%v=s, err=%v", string(debug.Stack()), err)
		}
	}()

	// 日志
	err := app.InitLog(*logFile, *outputConsole, map[string]interface{}{
		"server": "gateway",
	})
	if err != nil {
		panic(err)
	}

	// 配置文件
	cfg, err := config.LoadConfig(*configPath, *envPrefix)
	if err != nil {
		log.Errorf("load config fail. err=%v", err)
		os.Exit(1)
	}

	// consul注册中心
	registry.DefaultRegistry = consul_registry.NewRegistry(
		registry.Addrs(cfg.Registry.Addr),
	)

	if err = handler.WarmGrpcClient(cfg.WarmServer.Grpc); err != nil {
		log.Errorf("warm grpc client fail. err=%v", err)
		os.Exit(1)
	}

	svr := http_server.NewServer(&cfg.Server)
	if err := svr.Start(); err != nil {
		log.Errorf("start fail. port=%d, err=%v", cfg.Server.Port, err)
		return
	}

	defer svr.Stop()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case s := <-ch:
		log.Info("siganl for quit.", "sig", s)
	}
}

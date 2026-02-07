package grpc_server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/robert-pkg/base4go/log"
	"github.com/robert-pkg/base4go/registry"
	_ "github.com/robert-pkg/base4go/rpc/grpc/codec/json" //注册 json codec
	"github.com/robert-pkg/base4go/rpc/grpc/interceptor"
	"github.com/robert-pkg/base4go/rpc/server"
	net_utils "github.com/robert-pkg/base4go/utils/net"
)

type grpcServer struct {
	opts server.Options

	srv         *grpc.Server
	healthSrv   *health.Server
	metrics_srv *http.Server

	sync.RWMutex
	// marks the serve as started
	started bool
	// used for first registration
	registered bool

	host     string
	port     int
	httpPort int

	exit chan chan error

	// registry service instance
	reg_svc_map   map[string]*registry.Service
	registeredMap map[string]bool
}

func (g *grpcServer) Init() error {

	return nil
}

func (g *grpcServer) configure(opts ...server.Option) {
	g.Lock()
	defer g.Unlock()

	// Don't reprocess where there's no config
	if len(opts) == 0 && g.srv != nil {
		return
	}

	for _, o := range opts {
		o(&g.opts)
	}

	g.reg_svc_map = make(map[string]*registry.Service)

	gopts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.ServerRecoverInterceptor(),
			interceptor.ServerLogInterceptor(),
		),
	}

	if opts := g.getGrpcOptions(); opts != nil {
		gopts = append(gopts, opts...)
	}

	g.srv = grpc.NewServer(gopts...)
}

func (g *grpcServer) getGrpcOptions() []grpc.ServerOption {
	if g.opts.Context == nil {
		return nil
	}

	opts, ok := g.opts.Context.Value(grpcOptions{}).([]grpc.ServerOption)
	if !ok || opts == nil {
		return nil
	}

	return opts
}

func (g *grpcServer) startGrpcServer(port int) error {

	addr := ":" + strconv.Itoa(port)
	listen, err := net.Listen("tcp", addr) // hold port
	if err != nil {
		log.Errorf("GrpcServer net.Listen", "err", err)
		return err
	}

	log.Infof("startGrpcServer listen... addr=%s", addr)

	// 创建健康检查服务实例
	g.healthSrv = health.NewServer()

	// 注册到 gRPC Server. 这会自动处理 Check 和 Watch 请求
	healthpb.RegisterHealthServer(g.srv, g.healthSrv)

	// 服务不健康
	g.healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)

	reflection.Register(g.srv)

	go func() {
		if err = g.srv.Serve(listen); err != nil { // block
			log.Errorf("grpc server err: %v\r\n", err)
		}
	}()

	return nil
}

// 用作healthy 和 metrics
func (g *grpcServer) startHttpServer(port int) error {

	addr := fmt.Sprintf(":%d", port)
	log.Infof("startHttpServer listen... addr=%s", addr)

	router := gin.Default()

	router.GET("/status", func(c *gin.Context) {
		c.String(http.StatusOK, "status ok!")
	})

	g.metrics_srv = &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		err := g.metrics_srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("HttpServer err: %v\r\n", err)
		} else {
			log.Infof("HttpServer 正常关闭")
		}
	}()

	return nil
}

func (g *grpcServer) getServiceInfoList(opts *server.StartOptions) []*ServiceInfo {
	if opts.Context == nil {
		return []*ServiceInfo{}
	}

	if list, ok := opts.Context.Value(serviceInfoListKey{}).([]*ServiceInfo); ok && len(list) > 0 {
		return list
	}

	return []*ServiceInfo{}
}

func (g *grpcServer) Start(options ...server.StartOption) (err error) {
	g.RLock()
	if g.started {
		g.RUnlock()
		return nil
	}
	g.RUnlock()

	startOpts := server.StartOptions{
		Context: context.Background(),
	}

	for _, o := range options {
		o(&startOpts)
	}

	services := g.getServiceInfoList(&startOpts)
	if len(services) == 0 {
		return errors.New("no services")
	}

	g.host = g.opts.Host
	g.port = g.opts.Port
	g.httpPort = g.opts.HttpPort

	if g.port == 0 {
		if g.port, err = net_utils.RandPort(20000, 50000, 100); err != nil {
			return
		}
	}

	if g.httpPort == 0 {
		g.httpPort = g.port + 1
	}

	err = g.startHttpServer(g.httpPort)
	if err != nil {
		return err
	}

	for _, v := range services {
		v.RegisterOption(g.srv)
		g.registeredMap[v.GetKey()] = false
	}

	err = g.startGrpcServer(g.port)
	if err != nil {
		return err
	}

	g.buildRegService(services)

	defer func() {
		if err != nil {
			g.deregister()
		}
	}()
	for key, v := range g.reg_svc_map {
		if err := g.register(v); err != nil {
			return err
		} else {
			g.registeredMap[key] = true
		}
	}

	// 服务发现注册成功了。 则服务健康
	g.healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	log.Infof("grpc server设置为健康")

	go func() {
		t := time.NewTicker(time.Second * 10)
		defer func() {
			t.Stop()
		}()

		var ch chan error // return error chan
		isQuit := false
		for {
			if isQuit {
				break
			}
			select {
			// register self on interval
			case <-t.C:
				for _, v := range g.reg_svc_map {
					if err := g.register(v); err != nil {
						log.Errorf("register fail. %v\r\n", err)
					}
				}
				// wait for exit
			case ch = <-g.exit:
				isQuit = true
				break
			}
		}

		// deregister self
		if err := g.deregister(); err != nil {
			log.Errorf("Deregister fail err :%v \r\n", err)
		}

		// stop the grpc server
		exit := make(chan bool)

		go func() {
			g.srv.GracefulStop()
			close(exit)
		}()

		select {
		case <-exit:
		case <-time.After(time.Second * 10):
			log.Errorf("GracefulStop 超时, 则使用Stop()强制关闭")
			g.srv.Stop()
		}

		ch <- err
	}()

	// mark the server as started
	g.Lock()
	g.started = true
	g.Unlock()

	return nil
}

func (g *grpcServer) Stop() error {

	g.RLock()
	if !g.started {
		g.RUnlock()
		return nil
	}
	g.RUnlock()

	ch := make(chan error)
	g.exit <- ch

	var err error
	select {
	case err = <-ch:
		g.Lock()
		g.started = false
		g.Unlock()
	}

	return err
}

func (g *grpcServer) String() string {
	return "grpc"
}

func newGRPCServer(opts ...server.Option) *grpcServer {
	options := server.NewOptions(opts...)

	srv := &grpcServer{
		opts:          options,
		exit:          make(chan chan error),
		reg_svc_map:   make(map[string]*registry.Service),
		registeredMap: make(map[string]bool),
	}

	// configure the grpc server
	srv.configure()

	return srv
}

func NewServer(opts ...server.Option) server.Server {
	return newGRPCServer(opts...)
}

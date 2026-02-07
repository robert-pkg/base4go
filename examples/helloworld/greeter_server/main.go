package main

import (
	"flag"
	"os"

	"github.com/robert-pkg/base4go/app"
	pb "github.com/robert-pkg/base4go/examples/helloworld/greeter_server/api"
	"github.com/robert-pkg/base4go/rpc/server"
	"github.com/robert-pkg/base4go/rpc/server/grpc_server"
	"google.golang.org/grpc"
)

const (
	PackageName = "Greeter"
	ServieName  = "Greeter"
)

func main() {
	flag.Parse()

	a := app.New()
	if err := a.Init(); err != nil {
		os.Exit(1)
	}

	greeterServiceInfo := grpc_server.NewServiceInfo(PackageName, ServieName, "v1.0.0", func(s *grpc.Server) {
		pb.RegisterGreeterServer(s, &GreeterServerImpl{})
	})
	//greeterServiceInfo.SetServiceMetadata("desc", "这是一个greeter服务")
	//greeterServiceInfo.SetNodeMetadata("gray", "true")

	si := &app.ServerInfo{
		Protocol: "grpc",
		Host:     "127.0.0.1",
		StartOpts: []server.StartOption{
			grpc_server.ServiceInfoList([]*grpc_server.ServiceInfo{greeterServiceInfo}),
		},
	}

	a.Run(si)
}

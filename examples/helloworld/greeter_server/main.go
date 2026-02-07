package main

import (
	"context"
	"flag"

	"github.com/robert-pkg/base4go/app"
	pb "github.com/robert-pkg/base4go/examples/helloworld/greeter_server/api"
	"github.com/robert-pkg/base4go/log"
	"github.com/robert-pkg/base4go/rpc/server"
	"github.com/robert-pkg/base4go/rpc/server/grpc_server"
	"google.golang.org/grpc"
)

const (
	PackageName = "Greeter"
	ServieName  = "Greeter"
)

type GreeterServerImpl struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *GreeterServerImpl) SayHello(_ context.Context, in *pb.SayHelloRequest) (*pb.SayHelloReply, error) {
	log.Infof("Received: %v", in.GetName())
	return &pb.SayHelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	flag.Parse()

	a := app.New()

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

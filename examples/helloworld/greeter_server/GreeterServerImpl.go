package main

import (
	"context"

	"github.com/robert-pkg/base4go/app"
	pb "github.com/robert-pkg/base4go/examples/helloworld/greeter_server/api"
	"github.com/robert-pkg/base4go/log"
)

type GreeterServerImpl struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *GreeterServerImpl) SayHello(_ context.Context, in *pb.SayHelloRequest) (*pb.SayHelloReply, error) {
	log.Infof("Received: %v", in.GetName())

	if len(in.GetName()) < 3 {
		return app.Fail(&pb.SayHelloReply{}, 400, "name is too short"), nil

	}
	data := &pb.SayHelloReplyData{
		Message: in.GetName(),
	}

	rsp := app.Success(&pb.SayHelloReply{}, data)

	return rsp, nil
}

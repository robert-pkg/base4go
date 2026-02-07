package grpc_server

import (
	"google.golang.org/grpc"

	"github.com/robert-pkg/base4go/rpc/server"
)

type grpcOptions struct{}

// Options to be used to configure gRPC options.
func Options(opts ...grpc.ServerOption) server.Option {
	return setServerOption(grpcOptions{}, opts)
}

type serviceInfoListKey struct{}

func ServiceInfoList(services []*ServiceInfo) server.StartOption {
	return setStartOption(serviceInfoListKey{}, services)
}

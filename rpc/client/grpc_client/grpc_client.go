package grpc_client

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"

	//"github.com/robert-pkg/base4go/registry"
	"github.com/robert-pkg/base4go/log"
	"github.com/robert-pkg/base4go/rpc/client"
	"github.com/robert-pkg/base4go/rpc/grpc/balance"
	_ "github.com/robert-pkg/base4go/rpc/grpc/codec/json"
	"github.com/robert-pkg/base4go/rpc/grpc/interceptor"
	consul_resolver "github.com/robert-pkg/base4go/rpc/grpc/resolver/consul_resolver"
)

type grpcClient struct {
	Target string
	opts   client.Options

	conn *grpc.ClientConn
}

func (g *grpcClient) configure(opts ...client.Option) {
	for _, o := range opts {
		o(&g.opts)
	}

	consul_resolver.Register(g.opts.Registry)

}

func (g *grpcClient) Init() error {

	grpc_dial_opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 不使用 TLS（明文连接）
		grpc.WithDefaultServiceConfig(fmt.Sprintf(
			`{"loadBalancingConfig": [{"%s":{}}]}`, balance.BalancerName)),
		//grpc.WithNoProxy(), // 禁用代理，直接连接到后端
		grpc.WithDefaultCallOptions(grpc.ForceCodecV2(encoding.GetCodecV2("json"))),
		grpc.WithChainUnaryInterceptor(
			interceptor.ClientLogInterceptor(),
		),
	}

	conn, err := grpc.NewClient(g.Target, grpc_dial_opts...)
	if err != nil {
		log.Errorf("did not connect: %v", err)
		return err
	}

	g.conn = conn

	return nil
}

func (g *grpcClient) Invoke(ctx context.Context, method string, args, reply any) error {
	return g.conn.Invoke(ctx, method, args, reply)
}

func (g *grpcClient) Close() error {
	return g.conn.Close()
}

func (g *grpcClient) String() string {
	return "grpc"
}

func newGRPCClient(target string, opts ...client.Option) *grpcClient {
	options := client.NewOptions(opts...)

	srv := &grpcClient{
		Target: target,
		opts:   options,
	}

	// configure the grpc client
	srv.configure()

	return srv
}

func NewClient(target string, opts ...client.Option) client.Client {
	return newGRPCClient(target, opts...)
}

package handler

import (
	"github.com/robert-pkg/base4go/log"
	grpc_client "github.com/robert-pkg/base4go/rpc/client/grpc_client"
)

var (
	g_ClientMgr = grpc_client.GetClientMgr()
)

// 预热 grpc 客户端
func WarmGrpcClient(grpcClientTargets []string) error {
	if len(grpcClientTargets) > 0 {
		for _, target := range grpcClientTargets {
			_, err := g_ClientMgr.GetClient(target)
			if err != nil {
				log.Errorf("warm grpc client fail. target=%s, err=%v", target, err)
				return err
			}
		}
	}

	return nil
}

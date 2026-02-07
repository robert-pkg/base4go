package interceptor

import (
	"context"
	"runtime/debug"

	"google.golang.org/grpc"
	//grpc_metadata "google.golang.org/grpc/metadata"

	"github.com/robert-pkg/base4go/log"
)

func ServerRecoverInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Errorf("ServerRecoverInterceptor, err:%v stack:%v", e, string(debug.Stack()))
		}
	}()

	return handler(ctx, req)
}

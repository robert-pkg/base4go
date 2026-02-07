package interceptor

import (
	"context"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	//grpc_metadata "google.golang.org/grpc/metadata"

	"github.com/robert-pkg/base4go/log"
)

func ServerRecoverInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if e := recover(); e != nil {
			log.Errorf("ServerRecoverInterceptor, err:%v stack:%v", e, string(debug.Stack()))
		}

		return handler(ctx, req)
	}
}

func ServerLogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		start := time.Now()
		log.Debugf("inner log. method:%s req:%v", info.FullMethod, req)

		resp, err = handler(ctx, req)
		log.Debugf("inner log. method:%s duration:%v error:%v", info.FullMethod, time.Since(start), err)
		return resp, err
	}
}

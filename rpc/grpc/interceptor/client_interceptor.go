package interceptor

import (
	"context"
	"time"

	"github.com/robert-pkg/base4go/log"
	"google.golang.org/grpc"
)

func ClientLogInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, resp interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, resp, cc, opts...)

		log.Debug("outer log", "method", method, "target", cc.Target(), "duration", time.Since(start), "error", err)

		return err
	}
}

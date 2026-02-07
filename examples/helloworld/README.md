# helloworld

```
grpcurl -plaintext -d '{"name": "Robert"}' -format json localhost:29642 api.Greeter/SayHello
```

codec:

```

import (
	// 关键步骤：导入 json codec 以便客户端支持 json 编码
	_ "github.com/robert-pkg/base4go/rpc/grpc/codec/json"
)

// 方式二：JSON Codec 调用
// 传输内容：JSON 字符串 {"name": "Robert (JSON)"}
// Content-Type: application/grpc+json
func callWithJSON() {
	conn, err := grpc.Dial(*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// 全局指定：该连接的所有调用都使用 json codec
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype("json")),
	)
	if err != nil {
		log.Fatalf("JSON connect fail: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 也可以在单次调用中指定：grpc.CallContentSubtype("json")
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "Robert (JSON)"})
	if err != nil {
		log.Printf("JSON call fail: %v", err)
		return
	}
	log.Printf("[JSON Mode] Response: %s", r.GetMessage())
}
```

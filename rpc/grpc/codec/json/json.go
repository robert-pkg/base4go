package json

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/mem"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Name is the name registered for the proto compressor.
const Name = "json"

func init() {
	encoding.RegisterCodecV2(&codecV2{})
}

// codecV2 implements encoding.CodecV2.
type codecV2 struct{}

var (
	// marshalOptions 配置 protojson 序列化选项
	// EmitUnpopulated: true - 输出零值字段（如 0, ""），便于前端处理
	// UseProtoNames: true - 使用 proto 定义的字段名（通常是 snake_case），而非 Go 结构体的 CamelCase
	marshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
	}

	// unmarshalOptions 配置 protojson 反序列化选项
	// DiscardUnknown: true - 忽略 JSON 中存在但 Proto 定义中不存在的字段，避免报错
	unmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

func (c *codecV2) Marshal(v any) (data mem.BufferSlice, err error) {
	var b []byte
	// Check if the value is a proto message
	if m, ok := v.(proto.Message); ok {
		b, err = marshalOptions.Marshal(m)
	} else {
		// Fallback to standard json for non-proto messages
		b, err = json.Marshal(v)
	}

	if err != nil {
		return nil, err
	}

	// 封装为 gRPC 的 mem.BufferSlice
	// mem.NewBuffer 将 []byte 包装为 Buffer，nil 表示使用默认 pool
	return mem.BufferSlice{mem.NewBuffer(&b, nil)}, nil
}

func (c *codecV2) Unmarshal(data mem.BufferSlice, v any) (err error) {
	// 获取字节数据：Materialize() 将 BufferSlice 合并为 []byte
	b := data.Materialize()

	if m, ok := v.(proto.Message); ok {
		return unmarshalOptions.Unmarshal(b, m)
	}
	return json.Unmarshal(b, v)
}

func (c *codecV2) Name() string {
	return Name
}

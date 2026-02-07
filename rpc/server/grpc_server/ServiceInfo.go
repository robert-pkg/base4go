package grpc_server

import (
	"google.golang.org/grpc"
)

func NewServiceInfo(packageName, serviceName, version string, registerOption func(server *grpc.Server)) *ServiceInfo {
	return &ServiceInfo{
		PackageName:     packageName,
		ServiceName:     serviceName,
		Version:         version,
		serviceMetadata: map[string]string{},
		nodeMetadata: map[string]string{
			"protocol": "grpc",
			"language": "go",
		},
		RegisterOption: registerOption,
	}
}

type ServiceInfo struct {
	PackageName     string            // 包名
	ServiceName     string            // 服务名
	Version         string            // 服务版本
	serviceMetadata map[string]string // 服务元数据
	nodeMetadata    map[string]string // 节点元数据

	RegisterOption func(server *grpc.Server) // 注册服务到grpc
}

func (si *ServiceInfo) SetServiceMetadata(key, value string) {
	si.serviceMetadata[key] = value
}

func (si *ServiceInfo) SetNodeMetadata(key, value string) {
	si.nodeMetadata[key] = value
}

func (si *ServiceInfo) GetKey() string {
	return si.PackageName + "." + si.ServiceName
}

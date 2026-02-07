// Package registry is an interface for service discovery
package registry

import (
	"errors"
)

var (
	// Not found error when GetService is called.
	ErrNotFound = errors.New("service not found")
	// Watcher stopped error when watcher is stopped.
	ErrWatcherStopped = errors.New("watcher stopped")
)

// The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}.
type Registry interface {
	Init(...Option) error
	Options() Options
	Register(*Service, ...RegisterOption) error
	Deregister(*Service, ...DeregisterOption) error
	GetService(string, ...GetOption) ([]*Service, error)
	ListServices(...ListOption) ([]*Service, error)
	String() string
}

type Service struct {
	Name      string            `json:"name"`      // 服务名
	Version   string            `json:"version"`   // 服务版本
	Metadata  map[string]string `json:"metadata"`  // 服务级别的元数据
	Endpoints []*Endpoint       `json:"endpoints"` // 该服务暴露的接口（RPC 方法）列表
	Nodes     []*Node           `json:"nodes"`     // 当前这个服务有哪些实例在跑（IP+端口）
}

type Node struct {
	Id       string            `json:"id"`       // 实例唯一 ID
	Address  string            `json:"address"`  // 实例地址，比如 10.0.0.5:8080
	Metadata map[string]string `json:"metadata"` // 实例级别标签（机房、权重、版本等）
}

type Endpoint struct {
	Request  *Value            `json:"request"`
	Response *Value            `json:"response"`
	Metadata map[string]string `json:"metadata"`
	Name     string            `json:"name"`
}

type Value struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []*Value `json:"values"`
}

type Option func(*Options)

type RegisterOption func(*RegisterOptions)

type DeregisterOption func(*DeregisterOptions)

type GetOption func(*GetOptions)

type ListOption func(*ListOptions)

// Register a service node. Additionally supply options such as TTL.
func Register(s *Service, opts ...RegisterOption) error {
	return DefaultRegistry.Register(s, opts...)
}

// Deregister a service node.
func Deregister(s *Service) error {
	return DefaultRegistry.Deregister(s)
}

// Retrieve a service. A slice is returned since we separate Name/Version.
func GetService(name string, opts ...GetOption) ([]*Service, error) {
	return DefaultRegistry.GetService(name, opts...)
}

// List the services. Only returns service names.
func ListServices() ([]*Service, error) {
	return DefaultRegistry.ListServices()
}

func String() string {
	return DefaultRegistry.String()
}

var (
	DefaultRegistry Registry
)

package consul

import (
	"context"
	"fmt"
	"time"

	consul "github.com/hashicorp/consul/api"
	"github.com/robert-pkg/base4go/registry"
)

// Define a custom type for context keys to avoid collisions.
type contextKey string

const consulConfigKey contextKey = "consul_config"

// 注册接口的可选参数
const consulTCPCheckKey contextKey = "consul_tcp_check"
const consulHTTPCheckConfigKey contextKey = "consul_http_check_config"

// GetService的可选参数
const consulPassingOnlyKey contextKey = "consul_passing_only"
const consulQueryOptionsKey contextKey = "consul_query_options"
const consulGetServiceResultOptionsKey contextKey = "consul_get_service_result_options"

func Config(c *consul.Config) registry.Option {
	return func(o *registry.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, consulConfigKey, c)
	}
}

// TCPCheck will tell the service provider to check the service address
// and port every `t` interval. It will enabled only if `t` is greater than 0.
// See `TCP + Interval` for more information [1].
//
// [1] https://www.consul.io/docs/agent/checks.html
func TCPCheck(tcp string, interval, timeout time.Duration) registry.RegisterOption {
	return func(o *registry.RegisterOptions) {
		if interval <= time.Duration(0) || timeout <= time.Duration(0) {
			return
		}
		if o.Context == nil {
			o.Context = context.Background()
		}

		deregTTL := getDeregisterTTL(interval)
		check := consul.AgentServiceCheck{
			TCP:                            tcp,
			Interval:                       fmt.Sprintf("%v", interval),
			Timeout:                        fmt.Sprintf("%v", timeout),
			DeregisterCriticalServiceAfter: fmt.Sprintf("%v", deregTTL),
		}
		o.Context = context.WithValue(o.Context, consulTCPCheckKey, check)
	}
}

// HTTPCheck will tell the service provider to invoke the health check endpoint
// with an interval and timeout. It will be enabled only if interval and
// timeout are greater than 0.
// See `HTTP + Interval` for more information [1].
//
// [1] https://www.consul.io/docs/agent/checks.html
func HTTPCheck(http string, interval, timeout time.Duration) registry.RegisterOption {
	return func(o *registry.RegisterOptions) {
		if interval <= time.Duration(0) || timeout <= time.Duration(0) {
			return
		}
		if o.Context == nil {
			o.Context = context.Background()
		}

		deregTTL := getDeregisterTTL(interval)
		check := consul.AgentServiceCheck{
			HTTP:                           http,
			Interval:                       fmt.Sprintf("%v", interval),
			Timeout:                        fmt.Sprintf("%v", timeout),
			DeregisterCriticalServiceAfter: fmt.Sprintf("%v", deregTTL),
		}
		o.Context = context.WithValue(o.Context, consulHTTPCheckConfigKey, check)
	}
}

func PassingOnly(passingOnly bool) registry.GetOption {
	return func(o *registry.GetOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, consulPassingOnlyKey, passingOnly)
	}
}

// QueryOptions specifies the QueryOptions to be used when calling
// Consul. See `Consul API` for more information [1].
//
// [1] https://godoc.org/github.com/hashicorp/consul/api#QueryOptions
func QueryOptionsForGetService(q *consul.QueryOptions) registry.GetOption {
	return func(o *registry.GetOptions) {
		if q == nil {
			return
		}
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, consulQueryOptionsKey, q)
	}
}

func GetServiceResult(resultFn func(*consul.QueryMeta)) registry.GetOption {
	return func(o *registry.GetOptions) {
		if resultFn == nil {
			return
		}
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, consulGetServiceResultOptionsKey, resultFn)
	}
}

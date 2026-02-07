package balance

import (
	"math/rand"
	"sync/atomic"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	grpc_metadata "google.golang.org/grpc/metadata"

	"github.com/robert-pkg/base4go/log"
)

// 参考： github.com/grpc/grpc-go/balancer/roundrobin/roundrobin.go
const BalancerName = "z_round_robin"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(BalancerName, &builder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type builder struct{}

func (bb *builder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	scs := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for sc, _ := range info.ReadySCs {
		scs = append(scs, sc)
	}

	picker := &picker{subConns: scs}
	if len(scs) > 0 {
		// Start at a random index, as the same RR balancer rebuilds a new
		// picker when SubConn states change, and we don't want to apply excess
		// load to the first server in the list.
		picker.next = uint32(rand.Intn(len(scs)))
	}

	return picker
}

type picker struct {
	subConns []balancer.SubConn
	next     uint32
}

func (p *picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	ctx := info.Ctx
	md, ok := grpc_metadata.FromOutgoingContext(ctx)

	isGray := false
	if ok {
		if gray := md.Get("gray"); len(gray) > 0 && gray[0] == "true" {
			isGray = true
		}

		if isGray {
			log.Infof("灰度")
		}
	}

	if len(p.subConns) > 0 {
		subConnsLen := uint32(len(p.subConns))
		nextIndex := atomic.AddUint32(&p.next, 1)

		if nextIndex > 100000000 {
			atomic.StoreUint32(&p.next, 0)
		}
		sc := p.subConns[nextIndex%subConnsLen]
		return balancer.PickResult{SubConn: sc}, nil
	}

	return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
}

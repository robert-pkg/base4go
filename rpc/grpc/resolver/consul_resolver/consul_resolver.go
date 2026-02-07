package consul_resolver

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	consul_api "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"

	"github.com/robert-pkg/base4go/log"
	"github.com/robert-pkg/base4go/registry"
	consul_registry "github.com/robert-pkg/base4go/registry/consul"
)

func Register(r registry.Registry) {
	resolver.Register(&consulResolverBuilder{
		r: r,
	})
}

type consulResolverBuilder struct {
	r registry.Registry
}

func (*consulResolverBuilder) Scheme() string {
	return "consul"
}

func (b *consulResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {

	//log.Infof("consulResolverBuilder.Build enter. target:%s", target)
	trgt := target.URL.Host // parseTarget
	if trgt == "" {
		err := fmt.Errorf("url '%s' err. Must be in the next format: '%s://service{.dc}'", trgt, b.Scheme())
		log.Errorf("err:%v", err)
		return nil, err
	}

	var svc, dc string
	if true {
		split := strings.Split(trgt, ".")
		if len(split) == 1 {
			svc = split[0]
		} else if len(split) == 2 {
			svc, dc = split[0], split[1]
		} else {
			return nil, fmt.Errorf("url '%s' err. Must be in the next format: '%s://service{.dc}'", trgt, b.Scheme())
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	r := &consulResolver{
		registry: b.r,
		svc:      svc,
		dc:       dc,
		ctx:      ctx,
		cancel:   cancel,
		cc:       cc,

		addrMap: make(map[string]*registry.Service),
	}

	r.wg.Add(1)
	go r.watcher()
	return r, nil
}

// consulResolver implements resolver.Resolver
type consulResolver struct {
	registry registry.Registry
	svc, dc  string
	ctx      context.Context
	cancel   context.CancelFunc
	cc       resolver.ClientConn

	addrMap map[string]*registry.Service

	wg sync.WaitGroup
}

func (r *consulResolver) svcString() string {
	svc := r.svc
	if r.dc != "" {
		svc += "." + r.dc
	}

	return svc
}

func (r *consulResolver) watcher() {
	//log.Infof("consulResolver.watcher enter. dc:%s svc:%s", r.dc, r.svc)

	defer r.wg.Done()

	var lastIndex uint64
	errCnt := 0
	for {
		select {
		case <-r.ctx.Done():
			{
				log.Infof("consulResolver.watcher ctx.Done. so exit")
				return
			}
		default:
		}

		index, ss, err := r.getService(lastIndex)
		if err != nil {
			select {
			case <-r.ctx.Done():
				{
					//log.Infof("consulResolver.watcher ctx.Done. so exit")
					return
				}
			default:
			}

			log.Errorf("err: %v", err)
			errCnt += 1

			if errCnt >= 5 {
				time.Sleep(time.Minute)
			} else {
				time.Sleep(time.Duration(3*errCnt) * time.Second)
			}
			continue
		}

		errCnt = 0
		lastIndex = index
		newAddrMap := map[string]*registry.Service{}

		for _, svc := range ss {
			newAddrMap[svc.Nodes[0].Address] = svc
		}

		if len(newAddrMap) == 0 {
			if len(r.addrMap) > 0 {
				err := r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{{}}})
				if err != nil {
					log.Errorf("resolver state empty addr. watcher:%s", r.svcString(), "err", err)
					continue
				}

				r.addrMap = map[string]*registry.Service{}
				log.Infof("resolver state empty addr. watcher:%s", r.svcString(), "err", err)
			}

			continue
		} else {
			if len(newAddrMap) == len(r.addrMap) {
				isChange := false
				for key, _ := range newAddrMap {
					if _, ok := r.addrMap[key]; !ok {
						isChange = true
						break
					}
				}

				if !isChange {
					log.Info("no change")
					continue
				}
			}

			adds := make([]resolver.Address, 0, len(newAddrMap))
			for _, v := range newAddrMap {

				attr := attributes.New("version", v.Version)
				if len(v.Metadata) > 0 {
					attr = attr.WithValue("metadata", v.Metadata)
				}

				//log.Infof("service:%s version:%s, addr:%s, servie metadata:%v, node metadata:%v",v.Name, v.Version, v.Nodes[0].Address, v.Metadata, v.Nodes[0].Metadata)

				adds = append(adds, resolver.Address{
					Addr:       v.Nodes[0].Address,
					Attributes: attr,
				})
			}
			state := resolver.State{Addresses: adds}

			log.Infof("r.cc.UpdateState, count:%d , addrs:%v", len(state.Addresses), state.Addresses)
			err := r.cc.UpdateState(state)
			if err != nil {
				log.Errorf("resolver state update err. watcher:%s err:%v", r.svcString(), err)
				continue
			}
			r.addrMap = newAddrMap
		}
	}
}

func (r *consulResolver) getService(lastIndex uint64) (uint64, []*registry.Service, error) {
	ctx, cancel := context.WithTimeout(r.ctx, 60*time.Second)
	defer cancel()

	queryOptions := &consul_api.QueryOptions{
		WaitIndex:  lastIndex,
		Near:       "_agent",
		Datacenter: r.dc, // dc
		WaitTime:   time.Minute,
	}
	queryOptions = queryOptions.WithContext(ctx)

	var index uint64
	resultFn := func(queryMeta *consul_api.QueryMeta) {
		index = queryMeta.LastIndex
	}
	ss, err := r.registry.GetService(r.svc,
		consul_registry.PassingOnly(true),
		consul_registry.QueryOptionsForGetService(queryOptions),
		consul_registry.GetServiceResult(resultFn),
	)

	return index, ss, err
}

func (*consulResolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (r *consulResolver) Close() {
	//log.Infof("consulResolver close enter, %s", r.svcString())

	r.cancel()
	r.wg.Wait()

	//log.Infof("consulResolver close exit, %s", r.svcString())
}

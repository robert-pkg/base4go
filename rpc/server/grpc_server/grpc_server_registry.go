package grpc_server

import (
	"strconv"
	"time"

	"github.com/robert-pkg/base4go/log"
	"github.com/robert-pkg/base4go/registry"
	consul_registry "github.com/robert-pkg/base4go/registry/consul"
)

func (g *grpcServer) buildRegService(serviceInfoList []*ServiceInfo) {

	addr := g.host + ":" + strconv.Itoa(g.port)
	for _, servcieInfo := range serviceInfoList {
		// register service
		node := &registry.Node{
			Id:       servcieInfo.ServiceName + ":" + addr,
			Address:  addr,
			Metadata: servcieInfo.nodeMetadata,
		}

		reg_svc := &registry.Service{
			Name:      servcieInfo.ServiceName,
			Version:   servcieInfo.Version,
			Metadata:  servcieInfo.serviceMetadata,
			Nodes:     []*registry.Node{node},
			Endpoints: []*registry.Endpoint{},
		}

		log.Infof("Register service. key:%s node.Id:%s addr:%s", servcieInfo.GetKey(), node.Id, addr)
		g.reg_svc_map[servcieInfo.GetKey()] = reg_svc
	}
}

func (g *grpcServer) register(service *registry.Service) error {
	g.RLock()
	config := g.opts
	g.RUnlock()

	var regErr error
	for i := 0; i < 3; i++ {
		rOpts := []registry.RegisterOption{
			consul_registry.TCPCheck(service.Nodes[0].Address, time.Second*10, time.Second*5),
			consul_registry.HTTPCheck("http://"+g.host+":"+strconv.Itoa(g.httpPort)+"/status", time.Second*10, time.Second*5),
			registry.RegisterTTL(config.RegisterTTL),
		}

		// attempt to register
		if err := config.Registry.Register(service, rOpts...); err != nil {
			// set the error
			regErr = err
			// backoff then retry
			time.Sleep(time.Second * (1 << i))
			continue
		}

		// success so nil error
		regErr = nil
		break
	}

	return regErr

}

func (g *grpcServer) deregister() error {

	for key, isRegistered := range g.registeredMap {
		if isRegistered {
			if reg_svc, ok := g.reg_svc_map[key]; ok {
				if err := g.opts.Registry.Deregister(reg_svc); err != nil {
					log.Errorf("Deregister fail err :%v \r\n", err)
				} else {
					log.Infof("Deregister success. key:%s %v\r\n", key, reg_svc)
				}
			}
		}
	}

	return nil
}

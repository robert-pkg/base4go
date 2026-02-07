package grpc_client

import (
	"sync"

	"github.com/robert-pkg/base4go/log"
	"github.com/robert-pkg/base4go/rpc/client"
)

type ClientMgr interface {
	GetClient(target string) (client.Client, error)
}

func GetClientMgr() ClientMgr {
	return &clientMgr{
		clients: make(map[string]client.Client),
	}
}

type clientMgr struct {
	mu      sync.RWMutex
	clients map[string]client.Client
}

func (cm *clientMgr) GetClient(target string) (client.Client, error) {
	cm.mu.RLock()
	c := cm.clients[target]
	cm.mu.RUnlock()
	if c != nil {
		return c, nil
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	if c = cm.clients[target]; c != nil {
		return c, nil
	}

	cl := NewClient(target)
	if err := cl.Init(); err != nil {
		log.Infof("init client failed. target: %s, err: %v", target, err)
		return nil, err
	}

	cm.clients[target] = cl
	return cl, nil
}

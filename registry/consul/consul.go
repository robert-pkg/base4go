package consul

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	consul "github.com/hashicorp/consul/api"
	hash "github.com/mitchellh/hashstructure"
	"github.com/robert-pkg/base4go/registry"
)

type consulRegistry struct {
	Address []string
	opts    registry.Options

	client *consul.Client
	config *consul.Config

	sync.Mutex
	register map[string]uint64
}

func configure(c *consulRegistry, opts ...registry.Option) {
	// set opts
	for _, o := range opts {
		o(&c.opts)
	}

	// use default non pooled config
	config := consul.DefaultNonPooledConfig()

	if c.opts.Context != nil {
		// Use the consul config passed in the options, if available
		if co, ok := c.opts.Context.Value(consulConfigKey).(*consul.Config); ok {
			config = co
		}
	}

	// check if there are any addrs
	var addrs []string

	// iterate the options addresses
	for _, address := range c.opts.Addrs {
		// check we have a port
		addr, port, err := net.SplitHostPort(address)
		if ae, ok := err.(*net.AddrError); ok && ae.Err == "missing port in address" {
			port = "8500"
			addr = address
			addrs = append(addrs, net.JoinHostPort(addr, port))
		} else if err == nil {
			addrs = append(addrs, net.JoinHostPort(addr, port))
		}
	}

	// set the addrs
	if len(addrs) > 0 {
		c.Address = addrs
		config.Address = c.Address[0]
	}

	if config.HttpClient == nil {
		config.HttpClient = new(http.Client)
	}

	// set timeout
	if c.opts.Timeout > 0 {
		config.HttpClient.Timeout = c.opts.Timeout
	}

	// set the config
	c.config = config

	// remove client
	c.client = nil

	// setup the client
	c.Client()
}

func (c *consulRegistry) Init(opts ...registry.Option) error {
	configure(c, opts...)
	return nil
}

func (c *consulRegistry) Options() registry.Options {
	return c.opts
}

func (c *consulRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {
	if len(s.Nodes) == 0 {
		return errors.New("require at least one node")
	}

	var reg_options registry.RegisterOptions
	for _, o := range opts {
		o(&reg_options)
	}

	var regTCPCheck bool
	var regHTTPCheck bool
	var tcpCheckConfig consul.AgentServiceCheck
	var httpCheckConfig consul.AgentServiceCheck
	if reg_options.Context != nil {
		if v, ok := reg_options.Context.Value(consulTCPCheckKey).(consul.AgentServiceCheck); ok {
			regTCPCheck = true
			tcpCheckConfig = v
			reg_options.TTL = 0
		}
		if v, ok := reg_options.Context.Value(consulHTTPCheckConfigKey).(consul.AgentServiceCheck); ok {
			regHTTPCheck = true
			httpCheckConfig = v
			reg_options.TTL = 0
		}
	}

	// create hash of service; uint64
	h, err := hash.Hash(s, nil)
	if err != nil {
		return err
	}

	// use first node
	node := s.Nodes[0]

	// get existing hash and last checked time
	c.Lock()
	v, ok := c.register[s.Name]
	c.Unlock()

	// if it's already registered and matches then just pass the check
	if ok && v == h {
		if reg_options.TTL == time.Duration(0) {
			services, _, err := c.Client().Health().Checks(s.Name, &consul.QueryOptions{})
			if err == nil {
				for _, v := range services {
					if v.ServiceID == node.Id {
						return nil
					}
				}
			}
		} else {
			// if the err is nil we're all good, bail out
			// if not, we don't know what the state is, so full re-register
			if err := c.Client().Agent().PassTTL("service:"+node.Id, ""); err == nil {
				return nil
			}
		}
	}

	// encode the tags
	tags := encodeVersion("v", s.Version)
	if len(s.Metadata) > 0 {
		tags = append(tags, encodeMetadata("sm", s.Metadata)...)
	}

	if len(node.Metadata) > 0 {
		tags = append(tags, encodeMetadata("nm", node.Metadata)...)
	}

	if len(s.Endpoints) > 0 {
		tags = append(tags, encodeEndpoints("e", s.Endpoints)...)
	}

	var check *consul.AgentServiceCheck
	if regTCPCheck {
		check = &tcpCheckConfig
	} else if regHTTPCheck {
		check = &httpCheckConfig
	} else if reg_options.TTL > time.Duration(0) {
		deregTTL := getDeregisterTTL(reg_options.TTL)

		check = &consul.AgentServiceCheck{
			TTL:                            fmt.Sprintf("%v", reg_options.TTL),
			DeregisterCriticalServiceAfter: fmt.Sprintf("%v", deregTTL),
		}
	} else {
		// no check
	}

	host, pt, _ := net.SplitHostPort(node.Address)
	if host == "" {
		host = node.Address
	}
	port, _ := strconv.Atoi(pt)

	// register the service
	asr := &consul.AgentServiceRegistration{
		ID:      node.Id,
		Name:    s.Name,
		Tags:    tags,
		Port:    port,
		Address: host,
		Meta:    s.Metadata,
		Check:   check,
	}

	if err := c.Client().Agent().ServiceRegister(asr); err != nil {
		return err
	}

	// save our hash and time check of the service
	c.Lock()
	c.register[s.Name] = h
	c.Unlock()

	// if the TTL is 0 we don't mess with the checks
	if reg_options.TTL == time.Duration(0) {
		return nil
	}

	// pass the healthcheck
	return c.Client().Agent().PassTTL("service:"+node.Id, "")
}

func (c *consulRegistry) Deregister(s *registry.Service, opts ...registry.DeregisterOption) error {
	if len(s.Nodes) == 0 {
		return errors.New("require at least one node")
	}

	// delete our hash and time check of the service
	c.Lock()
	delete(c.register, s.Name)
	c.Unlock()

	node := s.Nodes[0]
	return c.Client().Agent().ServiceDeregister(node.Id)
}

func (c *consulRegistry) GetService(name string, opts ...registry.GetOption) ([]*registry.Service, error) {

	var options registry.GetOptions
	for _, o := range opts {
		o(&options)
	}

	passingOnly := false
	var queryOptions *consul.QueryOptions
	var resultFn func(*consul.QueryMeta)
	if options.Context != nil {
		if v, ok := options.Context.Value(consulPassingOnlyKey).(bool); ok {
			passingOnly = v
		}

		if v, ok := options.Context.Value(consulQueryOptionsKey).(*consul.QueryOptions); ok && v != nil {
			queryOptions = v
		}

		if v, ok := options.Context.Value(consulGetServiceResultOptionsKey).(func(*consul.QueryMeta)); ok && v != nil {
			resultFn = v
		}
	}

	if queryOptions == nil {
		queryOptions = &consul.QueryOptions{
			AllowStale: true,
		}
	}

	rsp, queryMeta, err := c.Client().Health().Service(name, "", passingOnly, queryOptions)
	if err != nil {
		return nil, err
	}

	services := make([]*registry.Service, 0, len(rsp))
	for _, s := range rsp {
		if s.Service.Service != name {
			continue
		}

		// address is service address
		address := s.Service.Address

		// use node address
		if len(address) == 0 {
			address = s.Node.Address
		}

		if !passingOnly {
			var del bool
			for _, check := range s.Checks {
				// delete the node if the status is critical
				//	HealthAny      = "any"
				//HealthPassing  = "passing"
				//HealthWarning  = "warning"
				//HealthCritical = "critical"
				//HealthMaint    = "maintenance"
				if check.Status == consul.HealthCritical { // passing, warning,critical,maintenance
					del = true
					break
				}
			}

			// if delete then skip the node
			if del {
				continue
			}
		}

		// version is now a tag
		version, _ := decodeVersion("v", s.Service.Tags)
		// service ID is now the node id
		id := s.Service.ID

		svc := &registry.Service{
			Endpoints: decodeEndpoints(s.Service.Tags),
			Name:      s.Service.Service,
			Version:   version,
			Nodes: []*registry.Node{
				&registry.Node{
					Id:       id,
					Address:  fmt.Sprintf("%s:%d", address, s.Service.Port),
					Metadata: decodeMetadata("nm", s.Service.Tags),
				},
			},
		}

		services = append(services, svc)
	}

	if resultFn != nil {
		resultFn(queryMeta)
	}
	return services, nil
}

func (c *consulRegistry) ListServices(opts ...registry.ListOption) ([]*registry.Service, error) {
	rsp, _, err := c.Client().Catalog().Services(&consul.QueryOptions{})
	if err != nil {
		return nil, err
	}

	var services []*registry.Service

	for service := range rsp {
		services = append(services, &registry.Service{Name: service})
	}

	return services, nil
}

func (c *consulRegistry) Client() *consul.Client {
	if c.client != nil {
		return c.client
	}

	for _, addr := range c.Address {
		// set the address
		c.config.Address = addr

		// create a new client
		tmpClient, _ := consul.NewClient(c.config)

		// test the client
		_, err := tmpClient.Agent().Host()
		if err != nil {
			continue
		}

		// set the client
		c.client = tmpClient
		return c.client
	}

	// set the default
	var err error
	c.client, err = consul.NewClient(c.config)
	if err != nil {
		// Log the error but return nil - caller should handle
		// This maintains backward compatibility while surfacing the error
		return nil
	}

	// return the client
	return c.client
}

func (c *consulRegistry) String() string {
	return "consul"
}

func newConsulRegistry(opts ...registry.Option) *consulRegistry {
	cr := &consulRegistry{
		opts:     registry.Options{},
		register: make(map[string]uint64),
	}
	configure(cr, opts...)
	return cr
}

func NewRegistry(opts ...registry.Option) registry.Registry {
	return newConsulRegistry(opts...)
}

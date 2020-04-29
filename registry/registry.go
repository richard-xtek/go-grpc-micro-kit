package registry

import (
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/consul/api"
	consul "github.com/hashicorp/consul/api"
)

// NewClient returns a new Client with connection to consul
func NewClient(addr string) (*consul.Client, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = addr

	c, err := consul.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// ConsulRegister ...
type ConsulRegister struct {
	ServiceName                    string   // service name
	Tags                           []string // consul tags
	ServicePort                    int      //service port
	DeregisterCriticalServiceAfter time.Duration
	Interval                       time.Duration
	Client                         *consul.Client
}

// NewConsulRegister ...
func NewConsulRegister(consulClient *consul.Client, serviceName string, servicePort int, tags []string) *ConsulRegister {
	return &ConsulRegister{
		ServiceName:                    serviceName,
		Tags:                           tags,
		ServicePort:                    servicePort,
		DeregisterCriticalServiceAfter: time.Duration(1) * time.Minute,
		Interval:                       time.Duration(10) * time.Second,
		Client:                         consulClient,
	}
}

// Register ...
func (r *ConsulRegister) Register() (string, error) {

	IP := localIP()
	reg := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%v-%v-%v", r.ServiceName, IP, r.ServicePort),
		Name:    r.ServiceName,
		Tags:    r.Tags,
		Port:    r.ServicePort,
		Address: IP,
		Check: &api.AgentServiceCheck{
			Interval:                       r.Interval.String(),
			GRPC:                           fmt.Sprintf("%v:%v/%v", IP, r.ServicePort, r.ServiceName),
			DeregisterCriticalServiceAfter: r.DeregisterCriticalServiceAfter.String(),
		},
	}
	return reg.ID, r.Client.Agent().ServiceRegister(reg)
}

// Deregister removes the service address from registry
func (r *ConsulRegister) Deregister(id string) error {
	return r.Client.Agent().ServiceDeregister(id)
}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

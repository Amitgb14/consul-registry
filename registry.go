package registry

import (
	"fmt"
	"os"
	"strings"

	consul "github.com/hashicorp/consul/api"
)

// Client client struct
type Client struct {
	consul   *consul.Client
	interval string
	timeout  string
}

// NewConsulClient create new consul client
func NewConsulClient(address, interval, timeout string) (*Client, error) {
	config := consul.DefaultConfig()
	config.Address = address
	c, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}
	return &Client{consul: c, interval: interval, timeout: timeout}, nil

}

// Register register service pn consul
func (c *Client) Register(name string, port int) error {
	agent := c.consul.Agent()
	hostname, _ := os.Hostname()
	shortname := strings.Split(hostname, ".")[0]
	id := fmt.Sprintf("%s:%s", name, shortname)
	service := fmt.Sprintf("%s:%v", shortname, port)
	reg := &consul.AgentServiceRegistration{
		ID:      id,
		Name:    name,
		Port:    port,
		Address: hostname,
		Check: &consul.AgentServiceCheck{
			Interval: c.interval,
			Timeout:  c.timeout,
			TCP:      service,
		},
	}
	agent.ServiceRegister(reg)
	return c.Service(name, service)
}

// DeRegister deregister from consul
func (c *Client) DeRegister(name string) error {
	agent := c.consul.Agent()
	hostname, _ := os.Hostname()
	shortname := strings.Split(hostname, ".")[0]
	id := fmt.Sprintf("%s:%s", name, shortname)
	err := agent.ServiceDeregister(id)
	if err != nil {
		return err
	}
	return nil
}

// Service setup service
func (c *Client) Service(name, tcp string) error {
	agent := c.consul.Agent()

	reg := &consul.AgentCheckRegistration{
		Name: name,
		AgentServiceCheck: consul.AgentServiceCheck{
			Status: consul.HealthPassing,
		},
	}

	reg.TCP = tcp
	reg.Interval = c.interval
	reg.Timeout = c.timeout
	err := agent.CheckRegister(reg)
	if err != nil {
		return err
	}

	checks, err := agent.Checks()
	if err != nil {
		return err
	}

	chk, ok := checks[name]
	if !ok {
		return fmt.Errorf("missing check: %v", ok)
	}

	if chk.Status != consul.HealthPassing {
		return fmt.Errorf("check not passing: %v", chk)
	}
	return nil
}

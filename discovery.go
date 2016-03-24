package egregor

import (
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-uuid"
)

type consulClient struct {
	id     string
	client *consul.Client
}

func (c *consulClient) Register(tags []string, port int) error {
	r := &consul.AgentServiceRegistration{
		ID:   c.id,
		Name: "command",
		Tags: tags,
		Port: port,
	}
	return c.client.Agent().ServiceRegister(r)
}

func (c *consulClient) Deregister() error {
	return c.client.Agent().ServiceDeregister(c.id)
}

func newConsulClient() (*consulClient, error) {
	c, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		return nil, err
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}

	return &consulClient{
		id:     id,
		client: c,
	}, nil
}

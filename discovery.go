package egregor

import (
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-uuid"
)

type consulClient struct {
	name   string
	id     string
	client *consul.Client
}

func (c *consulClient) Register(tags []string, port int) error {
	r := consul.AgentServiceRegistration{
		ID:   c.id,
		Name: c.name,
		Tags: tags,
		Port: port,
	}
	_ = r
	return nil
}

func newConsulClient(name string) (*consulClient, error) {
	c, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		return nil, err
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}

	return &consulClient{
		name:   name,
		id:     id,
		client: c,
	}, nil
}

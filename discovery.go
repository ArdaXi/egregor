package egregor

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-uuid"
)

type ConsulClient struct {
	id     string
	client *consul.Client
	sigint chan os.Signal
}

func (c *ConsulClient) Register(tags []string, port int) error {
	agent := c.client.Agent()

	r := &consul.AgentServiceRegistration{
		ID:   c.id,
		Name: "command",
		Tags: tags,
		Port: port,
		Check: &consul.AgentServiceCheck{
			TCP: (&net.TCPAddr{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: port,
			}).String(),
			Interval: "2s",
			Timeout:  "2s",
		},
	}

	signal.Notify(c.sigint, os.Interrupt)

	go func() {
		s := <-c.sigint
		log.Printf("Captured signal %s, deregistering and quitting.", s)
		c.Deregister()
		os.Exit(0)
	}()

	return agent.ServiceRegister(r)
}

func (c *ConsulClient) Deregister() error {
	return c.client.Agent().ServiceDeregister(c.id)
}

func (c *ConsulClient) GetKey(key string) ([]byte, error) {
	kv := c.client.KV()

	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, fmt.Errorf("Consul: %v not found", key)
	}

	return pair.Value, nil
}

func (c *ConsulClient) GetServiceAddr(command string) (string, error) {
	svcs, _, err := c.client.Catalog().Service("command", command, nil)
	if err != nil {
		return "", err
	}

	if len(svcs) == 0 {
		return "", fmt.Errorf("Nil response.")
	}

	svc := svcs[0]
	addr := svc.Address
	port := svc.ServicePort

	return fmt.Sprintf("%v:%v", addr, port), nil
}

func NewConsulClient() (*ConsulClient, error) {
	c, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		return nil, err
	}

	id, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}

	return &ConsulClient{
		id:     id,
		client: c,
		sigint: make(chan os.Signal, 1),
	}, nil
}

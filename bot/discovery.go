package main

import (
	"encoding/json"
	"fmt"

	consul "github.com/hashicorp/consul/api"
)

type consulClient struct {
	client *consul.Client
}

func NewConsulClient() (*consulClient, error) {
	c, err := consul.NewClient(consul.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return &consulClient{client: c}, nil
}

func (c *consulClient) getConfig() (*Config, error) {
	config := &Config{}
	kv := c.client.KV()

	pair, _, err := kv.Get("egregor/config", nil)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, fmt.Errorf("No config found")
	}

	err = json.Unmarshal(pair.Value, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *consulClient) getServiceAddr(command string) (string, error) {
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

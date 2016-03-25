package main

import (
	"encoding/json"
	"log"

	"github.com/ardaxi/egregor"
)

type Config struct {
	Server  string
	Name    string
	Channel string
}

func getConfig(consul *egregor.ConsulClient) (*Config, error) {
	config := &Config{}

	value, err := consul.GetKey("egregor/config")

	err = json.Unmarshal(value, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func main() {
	log.Println("Starting Consul client...")
	consul, err := egregor.NewConsulClient()
	if err != nil {
		log.Fatalf("Unexpected error creating Consul client: %v", err)
	}

	cfg, err := getConfig(consul)
	if err != nil {
		log.Fatalf("Unexpected error getting config from Consul: %v", err)
	}

	log.Printf("Connecting to %v as %v and joining %v", cfg.Server, cfg.Name, cfg.Channel)
	bot := NewBot(cfg.Server, cfg.Name, cfg.Channel, consul)
	err = bot.Connect()
	if err != nil {
		log.Fatalf("Bot error: %v", err)
	}

	bot.HandleLoop()
}

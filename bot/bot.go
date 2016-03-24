package main

import "log"

type Config struct {
	Server  string
	Name    string
	Channel string
}

func main() {
	log.Println("Starting Consul client...")
	consul, err := NewConsulClient()
	if err != nil {
		log.Fatalf("Unexpected error creating Consul client: %v", err)
	}

	cfg, err := consul.getConfig()
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

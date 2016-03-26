package main

import (
	"encoding/json"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/ardaxi/egregor"
	"github.com/ardaxi/egregor/pb"
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

func handleLoop(consul *egregor.ConsulClient, b *Bot) error {
	loggers, err := consul.GetLoggers()
	if err != nil {
		return err
	}

	for _, logger := range loggers {
		conn, err := grpc.Dial(logger, grpc.WithInsecure())
		if err != nil {
			return err
		}

		client := pb.NewCommandClient(conn)

		stream, err := client.StreamLog(context.Background())
		if err != nil {
			return err
		}

		go func() {
			// TODO: b.log should be copied for multiple loggers
			for msg := range b.log {
				if err := stream.Send(msg); err != nil {
					break
				}
			}

			conn.Close()
		}()
	}

	return nil
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

	err = handleLoop(consul, bot)
	if err != nil {
		log.Fatalf("Failed logger loop: %v", err)
	}

	bot.HandleLoop()
}

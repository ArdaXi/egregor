package main

import (
	"log"

	"google.golang.org/grpc"

	"github.com/ardaxi/egregor"
	"github.com/ardaxi/egregor/pb"
	"golang.org/x/net/context"
)

func main() {
	s, err := egregor.NewServer(&egregor.Config{})
	if err != nil {
		log.Fatalf("Egregor error: %v", err)
	}

	usage := []string{
		"help [command] returns usage information for a command, or a list of commands if command is omitted.",
	}

	s.HandleFunc("help", usage, makeHelpHandler(s))
	s.Run()
}

func makeHelpHandler(s egregor.Server) func(context.Context,
	*pb.CommandRequest) (*pb.CommandResponse, error) {
	return func(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
		if len(req.Args) == 0 {
			return listHandler(ctx, s)
		}

		command := req.Args[0]
		return usageHandler(ctx, s, command)
	}
}

func listHandler(ctx context.Context, s egregor.Server) (*pb.CommandResponse, error) {
	return &pb.CommandResponse{
		Reply: []string{"Not implemented yet."},
	}, nil
}

func usageHandler(ctx context.Context, s egregor.Server, command string) (*pb.CommandResponse, error) {
	addr, err := s.Consul().GetServiceAddr(command)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	c := pb.NewCommandClient(conn)

	return c.GetUsage(ctx, &pb.HelpRequest{Command: command})
}

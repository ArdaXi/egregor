package main

import (
	"fmt"
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

func isIn(m map[string]bool, vs []string) bool {
	for _, v := range vs {
		if !m[v] {
			return false
		}
	}
	return true
}

func listHandler(ctx context.Context, s egregor.Server) (*pb.CommandResponse, error) {
	svcs, err := s.Consul().GetCommands()
	if err != nil {
		return nil, err
	}

	got := make(map[string]bool)
	usage := make(map[string]string)

	for addr, tags := range svcs {
		if isIn(got, tags) {
			continue
		}

		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			continue
		}

		c := pb.NewCommandClient(conn)
		u, err := c.GetCommands(ctx, &pb.Empty{})
		if err != nil {
			continue
		}

		for _, entry := range u.List {
			usage[entry.Command] = entry.Description
		}
		conn.Close()

		for _, t := range tags {
			got[t] = true
		}
	}

	reply := []string{}
	for cmd, desc := range usage {
		reply = append(reply, fmt.Sprintf("%v: %v", cmd, desc))
	}

	return &pb.CommandResponse{
		Reply: reply,
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

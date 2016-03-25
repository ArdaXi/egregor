package egregor

//go:generate protoc -I ./pb ./pb/egregor.proto --go_out=plugins=grpc:pb

import (
	"fmt"
	"net"

	"github.com/ardaxi/egregor/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// A CommandHandler responds to a command
type CommandHandler interface {
	Handle(context.Context, *pb.CommandRequest) (*pb.CommandResponse, error)
	Usage() []string
}

type defaultHandler struct {
	handler func(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error)
	usage   []string
}

func (h *defaultHandler) Handle(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
	return h.handler(ctx, req)
}

func (h *defaultHandler) Usage() []string {
	return h.usage
}

type commandServer struct {
	handlers map[string]CommandHandler
}

func (s *commandServer) DoCommand(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
	cmd := req.Command
	handler, ok := s.handlers[cmd]
	if !ok {
		return nil, fmt.Errorf("Unknown command: %s", cmd)
	}

	return handler.Handle(ctx, req)
}

func (s *commandServer) GetUsage(ctx context.Context, req *pb.HelpRequest) (*pb.CommandResponse, error) {
	cmd := req.Command
	handler, ok := s.handlers[cmd]
	if !ok {
		return nil, fmt.Errorf("Unknown command: %s", cmd)
	}

	return &pb.CommandResponse{Reply: handler.Usage()}, nil
}

func (s *commandServer) GetCommands(ctx context.Context, req *pb.Empty) (*pb.CommandList, error) {
	entries := []*pb.CommandEntry{}
	for command, handler := range s.handlers {
		entry := &pb.CommandEntry{
			Command:     command,
			Description: handler.Usage()[0],
		}
		entries = append(entries, entry)
	}

	return &pb.CommandList{List: entries}, nil
}

func newCommandServer() *commandServer {
	return &commandServer{handlers: make(map[string]CommandHandler)}
}

type Server interface {
	Consul() *ConsulClient
	GetKey(key string) ([]byte, error)
	Handle(command string, handler CommandHandler)
	HandleFunc(command string, usage []string, handler func(context.Context, *pb.CommandRequest) (*pb.CommandResponse, error))
	Run() error
}

type server struct {
	consul        *ConsulClient
	commandServer *commandServer
	grpcServer    *grpc.Server
	config        *Config
}

type Config struct {
}

// NewServer initializes a gRPC server with the given config.
func NewServer(cfg *Config) (Server, error) {
	c, err := NewConsulClient()
	if err != nil {
		return nil, err
	}

	s := &server{
		commandServer: newCommandServer(),
		grpcServer:    grpc.NewServer(),
		config:        cfg,
		consul:        c,
	}

	pb.RegisterCommandServer(s.grpcServer, s.commandServer)

	return s, nil
}

func (s *server) getCommands() (keys []string) {
	for k := range s.commandServer.handlers {
		keys = append(keys, k)
	}
	return
}

// Consul returns the ConsulClient for this server
func (s *server) Consul() *ConsulClient {
	return s.consul
}

// GetKey retrieves a key from the Consul Key/Value store.
func (s *server) GetKey(key string) ([]byte, error) {
	return s.consul.GetKey(key)
}

// Handle registers the handler for the given command.
func (s *server) Handle(command string, handler CommandHandler) {
	s.commandServer.handlers[command] = handler
}

// HandleFunc registers the handler function for the given command.
func (s *server) HandleFunc(command string, usage []string,
	handler func(context.Context, *pb.CommandRequest) (*pb.CommandResponse, error)) {
	s.commandServer.handlers[command] = &defaultHandler{handler: handler, usage: usage}
}

// Run allocates a free TCP port, registers the commands in Consul and starts the gRPC service.
func (s *server) Run() error {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}

	commands := s.getCommands()
	port := ln.Addr().(*net.TCPAddr).Port
	if err := s.consul.Register(commands, port); err != nil {
		return err
	}
	defer s.consul.Deregister()

	return s.grpcServer.Serve(ln)
}

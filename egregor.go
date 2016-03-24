package egregor

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
}

type defaultHandler struct {
	handler func(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error)
}

func (h *defaultHandler) Handle(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
	return h.handler(ctx, req)
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

func newCommandServer() *commandServer {
	return &commandServer{handlers: make(map[string]CommandHandler)}
}

type Server interface {
	Handle(command string, handler CommandHandler)
	HandleFunc(command string, handler func(context.Context, *pb.CommandRequest) (*pb.CommandResponse, error))
	Run() error
}

type server struct {
	commandServer *commandServer
	grpcServer    *grpc.Server
	config        *Config
	consul        *consulClient
}

type Config struct {
	Name string
}

// NewServer initializes a gRPC server with the given config.
func NewServer(cfg *Config) (Server, error) {
	c, err := newConsulClient(cfg.Name)
	if err != nil {
		return nil, err
	}

	s := &server{
		commandServer: &commandServer{},
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

// Handle registers the handler for the given command.
func (s *server) Handle(command string, handler CommandHandler) {
	s.commandServer.handlers[command] = handler
}

// HandleFunc registers the handler function for the given command.
func (s *server) HandleFunc(command string,
	handler func(context.Context, *pb.CommandRequest) (*pb.CommandResponse, error)) {
	s.commandServer.handlers[command] = &defaultHandler{handler: handler}
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

	return s.grpcServer.Serve(ln)
}

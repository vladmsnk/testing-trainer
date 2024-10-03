package grpc_server

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
)

type Config struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Server struct {
	Ser  *grpc.Server
	Addr string
	lis  net.Listener
}

func NewGRPCServer(cfg *Config, opts ...grpc.ServerOption) (*Server, error) {
	c := &Server{
		Ser:  grpc.NewServer(opts...),
		Addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}

	var err error
	if c.lis, err = net.Listen("tcp", c.Addr); err != nil {
		return nil, err
	}

	return c, nil
}

func (ser *Server) Run() {
	go func() {
		_ = ser.Ser.Serve(ser.lis)
	}()
}

func (ser *Server) Close() {
	if ser.lis == nil {
		return
	}
	ser.Ser.GracefulStop()
}

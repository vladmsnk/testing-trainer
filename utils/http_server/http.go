package http_server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"google.golang.org/protobuf/proto"
)

type Config struct {
	Method         string        `yaml:"method"`
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	ReadTimeout    time.Duration `yaml:"readTimeout"`    // Время ожидания web запроса в секундах
	WriteTimeout   time.Duration `yaml:"writeTimeout"`   // Время ожидания окончания передачи ответа в секундах
	IdleTimeout    time.Duration `yaml:"idleTimeout"`    // Время ожидания следующего запроса
	MaxHeaderBytes int           `yaml:"maxHeaderBytes"` // Максимальный размер заголовка получаемого от браузера клиента в байтах
}

type Server struct {
	server  *http.Server
	lis     net.Listener
	control chan struct{}
}

func New(cfg *Config, handler http.Handler) (s *Server, err error) {
	s = &Server{
		server: &http.Server{
			Addr:           fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			Handler:        handler,
			ReadTimeout:    cfg.ReadTimeout,
			WriteTimeout:   cfg.WriteTimeout,
			IdleTimeout:    cfg.IdleTimeout,
			MaxHeaderBytes: cfg.MaxHeaderBytes,
		},
		control: make(chan struct{}),
	}

	if s.lis, err = net.Listen("tcp", s.server.Addr); err != nil {
		return nil, err
	}

	go func() {
		_ = s.server.Serve(s.lis)
		close(s.control)
	}()
	return s, nil
}

func (s *Server) Close() error {
	if s.lis == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		if err := s.lis.Close(); err != nil {
			return err
		}
	}
	<-s.control
	return nil
}

func ResponseHeaderMatcher(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	headers := w.Header()
	if location, ok := headers["Grpc-Metadata-Location"]; ok {
		w.Header().Set("Location", location[0])
		w.WriteHeader(http.StatusFound)
	}

	return nil
}

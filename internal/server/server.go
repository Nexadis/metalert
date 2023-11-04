// server для сбора метрик
package server

import (
	"context"

	"github.com/Nexadis/metalert/internal/storage"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	h *httpServer
	g *grpcServer
}

// Run Запуск сервера
func (s *Server) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return s.h.Run(ctx)
	})
	group.Go(func() error {
		return s.g.Run(ctx)
	})

	return group.Wait()
}

// New Конструктор Server, для инциализации использует Config
func New(config *Config) (*Server, error) {
	var err error
	if config == nil {
		config = NewConfig()
		config.SetDefault()
	}
	storage, err := storage.ChooseStorage(context.Background(), config.DB)
	if err != nil {
		return nil, err
	}
	httpserver, err := NewHTTPServer(config, storage)
	if err != nil {
		return nil, err
	}

	grpcserver, err := NewGRPCServer(config, storage)
	if err != nil {
		return nil, err
	}
	server := Server{
		httpserver,
		grpcserver,
	}
	return &server, nil
}

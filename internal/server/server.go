// server для сбора метрик
package server

import (
	"context"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Nexadis/metalert/internal/server/middlewares"
	"github.com/Nexadis/metalert/internal/storage"
	"github.com/Nexadis/metalert/internal/utils/asymcrypt"
)

// HTTPServer связывает все обработчики с базой данных
type HTTPServer struct {
	router     http.Handler
	storage    storage.Storage
	config     *Config
	privKey    []byte
	trustedNet *net.IPNet
}

type Server struct {
	h *HTTPServer
	g *GRPCServer
}

// Run Запуск сервера
func (s *Server) Run(ctx context.Context) error {
	storage, err := storage.ChooseStorage(context.Background(), s.h.config.DB)
	if err != nil {
		return err
	}
	s.h.storage = storage
	s.g.storage = storage
	l, err := net.Listen("tcp", s.h.config.Address)
	if err != nil {
		return err
	}
	defer l.Close()
	go func() error {
		err = http.Serve(l, s.h.router)
		return err
	}()

	<-ctx.Done()
	return err
}

// New Конструктор Server, для инциализации использует Config
func New(config *Config) (*Server, error) {
	var err error
	if config == nil {
		config = NewConfig()
		config.SetDefault()
	}
	var key []byte
	if config.CryptoKey != "" {
		key, err = asymcrypt.ReadPem(config.CryptoKey)
		if err != nil {
			return nil, err
		}
	}
	var trusted *net.IPNet = nil
	if config.TrustedSubnet != "" {
		_, trusted, err = net.ParseCIDR(config.TrustedSubnet)
		if err != nil {
			return nil, err
		}
	}
	httpserver := &HTTPServer{
		nil,
		nil,
		config,
		key,
		trusted,
	}
	httpserver.MountHandlers()
	grpcserver := &GRPCServer{
		config: config,
	}
	server := Server{
		httpserver,
		grpcserver,
	}
	return &server, nil
}

// MountHandlers Подключает все обработчики и middlewares к роутеру
func (s *HTTPServer) MountHandlers() {
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Get("/", s.InfoPage)
		r.Post("/updates/", s.Updates)
		r.Route("/update", func(r chi.Router) {
			r.Post("/", s.UpdateJSON)
			r.Post("/{mtype}/{id}/{value}", s.Update)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/", s.Values)
			r.Post("/", s.ValueJSON)
			r.Get("/{mtype}/{id}", s.Value)
		})
		r.Get("/ping", s.DBPing)
	})

	s.router = middlewares.WithTrusted(
		middlewares.WithDeflate(
			middlewares.WithDecrypt(
				middlewares.WithLogging(
					middlewares.WithVerify(
						router,
						s.config.SignKey,
					),
				),
				s.privKey,
			),
		),
		s.trustedNet,
	)
}

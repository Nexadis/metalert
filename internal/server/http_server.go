package server

import (
	"context"
	"net"
	"net/http"

	"github.com/Nexadis/metalert/internal/server/middlewares"
	"github.com/Nexadis/metalert/internal/storage"
	"github.com/Nexadis/metalert/internal/utils/asymcrypt"
	"github.com/Nexadis/metalert/internal/utils/logger"
	"github.com/go-chi/chi/v5"
)

// httpServer связывает все обработчики с базой данных
type httpServer struct {
	router     http.Handler
	storage    storage.Storage
	config     *Config
	privKey    []byte
	trustedNet *net.IPNet
}

func NewHTTPServer(config *Config, storage storage.Storage) (*httpServer, error) {
	var err error
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
	httpserver := &httpServer{
		nil,
		storage,
		config,
		key,
		trusted,
	}
	httpserver.MountHandlers()
	return httpserver, nil
}

// MountHandlers Подключает все обработчики и middlewares к роутеру
func (s *httpServer) MountHandlers() {
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

func (s *httpServer) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return err
	}
	defer l.Close()
	go func() {
		logger.Info("HTTP server at ", s.config.Address)
		err = http.Serve(l, s.router)
	}()
	<-ctx.Done()
	if err != nil {
		return err
	}
	return l.Close()
}

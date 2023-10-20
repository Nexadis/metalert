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
	router  http.Handler
	storage storage.Storage
	config  *Config
	privKey []byte
}

// Run Запуск сервера
func (s *HTTPServer) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return err
	}
	defer l.Close()
	go func() error {
		err = http.Serve(l, s.router)
		return err
	}()

	<-ctx.Done()
	return err
}

// NewServer Конструктор HTTPServer, для инциализации использует Config
func NewServer(config *Config) (*HTTPServer, error) {
	if config == nil {
		config = NewConfig()
		config.SetDefault()
	}
	storage, err := storage.ChooseStorage(context.Background(), config.DB)
	if err != nil {
		return nil, err
	}
	var key []byte
	if config.CryptoKey != "" {
		key, err = asymcrypt.ReadPem(config.CryptoKey)
		if err != nil {
			return nil, err
		}
	}
	server := &HTTPServer{
		nil,
		storage,
		config,
		key,
	}
	return server, nil
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

	s.router = middlewares.WithDeflate(
		middlewares.WithDecrypt(
			middlewares.WithLogging(
				middlewares.WithVerify(
					router,
					s.config.SignKey,
				),
			),
			s.privKey,
		),
	)
}

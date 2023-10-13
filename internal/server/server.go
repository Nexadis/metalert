// server для сбора метрик
package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Nexadis/metalert/internal/server/middlewares"
	"github.com/Nexadis/metalert/internal/storage"
	"github.com/Nexadis/metalert/internal/storage/db"
	"github.com/Nexadis/metalert/internal/storage/mem"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

type Listener interface {
	Run() error
	MountHandlers()
}

// HTTPServer связывает все обработчики с базой данных
type HTTPServer struct {
	router  http.Handler
	storage storage.Storage
	config  *Config
}

// Run Запуск сервера
func (s *HTTPServer) Run(ctx context.Context) error {
	storage, err := chooseStorage(ctx, s.config)
	if err != nil {
		return err
	}
	s.storage = storage
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

// chooseStorage Определяет по конфигу какое хранилище использовать
func chooseStorage(ctx context.Context, config *Config) (storage.Storage, error) {
	if config.DB.DSN != "" {
		logger.Info("Start with DB")
		p := db.New(config.DB)
		dbctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
		defer cancel()
		err := p.Open(dbctx, config.DB.DSN)
		if err == nil {
			return p, nil
		}
		logger.Error(err)
	}
	return getMemStorage(ctx, config)
}

func getMemStorage(ctx context.Context, config *Config) (mem.MetricsStorage, error) {
	logger.Info("Use in mem storage")
	metricsStorage := mem.NewMetricsStorage()
	err := metricsStorage.Restore(ctx, config.FileStoragePath, config.Restore)
	if err != nil {
		logger.Info(err)
		return nil, err
	}
	go metricsStorage.SaveTimer(ctx, config.FileStoragePath, config.StoreInterval)
	return metricsStorage, nil
}

// NewServer Конструктор HTTPServer, для инциализации использует Config
func NewServer(config *Config) (*HTTPServer, error) {
	if config == nil {
		config = NewConfig()
		config.SetDefault()
	}
	server := &HTTPServer{
		nil,
		nil,
		config,
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
		middlewares.WithLogging(
			s.WithVerify(
				router),
		),
	)
}

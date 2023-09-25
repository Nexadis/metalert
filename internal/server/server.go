// server для сбора метрик
package server

import (
	"context"
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

// Запуск сервера
func (s *HTTPServer) Run() error {
	return http.ListenAndServe(s.config.Address, s.router)
}

// Определяет по конфигу какое хранилище использовать
func chooseStorage(config *Config) (storage.Storage, error) {
	ctx := context.TODO()
	switch {
	case config.DB.DSN != "":
		logger.Info("Start with DB")
		db := db.New(config.DB)
		dbctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
		defer cancel()
		err := db.Open(dbctx, config.DB.DSN)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		err = db.Ping()
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		return db, nil
	default:
		logger.Info("Use in mem storage")
		metricsStorage := mem.NewMetricsStorage()
		err := metricsStorage.Restore(ctx, config.FileStoragePath, config.Restore)
		if err != nil {
			logger.Info(err)
			return nil, err
		}
		go metricsStorage.SaveTimer(context.Background(), config.FileStoragePath, config.StoreInterval)
		return metricsStorage, nil
	}
}

// Конструктор HTTPServer, для инциализации использует Config
func NewServer(config *Config) (*HTTPServer, error) {
	storage, err := chooseStorage(config)
	if err != nil {
		return nil, err
	}
	server := &HTTPServer{
		nil,
		storage,
		config,
	}
	return server, nil
}

// Подключает все обработчики и middlewares к роутеру
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

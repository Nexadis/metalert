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

type httpServer struct {
	router  http.Handler
	storage storage.Storage
	config  *Config
}

func (s *httpServer) Run() error {
	return http.ListenAndServe(s.config.Address, s.router)
}

func chooseStorage(config *Config) storage.Storage {
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
		}
		return db
	default:
		logger.Info("Use in mem storage")
		metricsStorage := mem.NewMetricsStorage()
		err := metricsStorage.Restore(ctx, config.FileStoragePath, config.Restore)
		if err != nil {
			logger.Info(err)
		}
		go metricsStorage.SaveTimer(context.Background(), config.FileStoragePath, config.StoreInterval)
		return metricsStorage
	}
}

func NewServer(config *Config) Listener {
	storage := chooseStorage(config)
	server := &httpServer{
		nil,
		storage,
		config,
	}
	return server
}

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
	s.router = middlewares.WithDeflate(middlewares.WithLogging(router))
}

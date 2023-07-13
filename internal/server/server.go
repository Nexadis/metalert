package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	saver   mem.StateSaver
	db      db.DataBase
	config  *Config
	exit    chan os.Signal
}

func (s *httpServer) Run() error {
	if s.db != nil {
		return http.ListenAndServe(s.config.Address, s.router)
	}
	go http.ListenAndServe(s.config.Address, s.router)
	go s.saver.SaveTimer(s.config.FileStoragePath, s.config.StoreInterval)
	for {
		<-s.exit
		s.Shutdown()
		return nil
	}
}

func NewServer(config *Config) Listener {
	switch {
	case config.DB.DSN != "":
		logger.Info("Start with DB")

		db := db.NewDB()
		dbctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second))
		defer cancel()
		err := db.Open(dbctx, config.DB.DSN)
		if err != nil {
			logger.Error(err)
		}
		server := &httpServer{
			nil,
			db,
			nil,
			db,
			config,
			nil,
		}
		return server
	default:
		logger.Info("Use in mem storage")
	}
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM|syscall.SIGINT|syscall.SIGQUIT)
	metricsStorage := mem.NewMetricsStorage()
	server := &httpServer{
		nil,
		metricsStorage,
		metricsStorage,
		nil,
		config,
		exit,
	}
	err := server.saver.Restore(server.config.FileStoragePath, server.config.Restore)
	if err != nil {
		logger.Info(err)
	}
	return server
}

func (s *httpServer) MountHandlers() {
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Get("/", s.InfoPage)
		r.Route("/update", func(r chi.Router) {
			r.Post("/", s.UpdateJSONHandler)
			r.Post("/{valType}/{name}/{value}", s.UpdateHandler)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/", s.ValuesHandler)
			r.Post("/", s.ValueJSONHandler)
			r.Get("/{valType}/{name}", s.ValueHandler)
		})
		r.Get("/ping", s.DBPing)
	})
	s.router = middlewares.WithDeflate(middlewares.WithLogging(router))
}

func (s *httpServer) Shutdown() {
	if s.db != nil {
		s.db.Close()
		return
	}
	s.saver.Save(s.config.FileStoragePath)
}

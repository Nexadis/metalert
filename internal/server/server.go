package server

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"

	"github.com/Nexadis/metalert/internal/server/middlewares"
	"github.com/Nexadis/metalert/internal/storage"
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
	exit    chan os.Signal
}

func (s *httpServer) Run() error {
	go s.storage.SaveTimer(s.config.FileStoragePath, s.config.StoreInterval)
	go http.ListenAndServe(s.config.Address, s.router)
	for {
		<-s.exit
		s.Shutdown()
		return nil
	}
}

func NewServer(config *Config) Listener {
	metricsStorage := storage.NewMetricsStorage()
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM|syscall.SIGINT|syscall.SIGQUIT)
	server := &httpServer{
		nil,
		metricsStorage,
		config,
		exit,
	}
	err := server.storage.Restore(server.config.FileStoragePath, server.config.Restore)
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
	})
	s.router = middlewares.WithDeflate(middlewares.WithLogging(router))
}

func (s *httpServer) Shutdown() {
	s.storage.Save(s.config.FileStoragePath)
}

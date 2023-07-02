package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/server/middlewares"
)

type Listener interface {
	Run() error
	MountHandlers()
}

type httpServer struct {
	router  http.Handler
	storage metrx.MemStorage
	config  *Config
	exit    chan os.Signal
}

func (s *httpServer) Run() error {
	go http.ListenAndServe(s.config.Address, s.router)
	if s.config.StoreInterval <= 0 {
		s.config.StoreInterval = 1
	}
	ticker := time.NewTicker(time.Duration(s.config.StoreInterval) * time.Second)
	for {
		select {
		case <-ticker.C:
			err := saveStorage(s)
			if err != nil {
				return err
			}

		case <-s.exit:
			s.Shutdown()
			return nil
		}
	}
}

func NewServer(config *Config) Listener {
	metricsStorage := metrx.NewMetricsStorage()
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	server := &httpServer{
		nil,
		metricsStorage,
		config,
		exit,
	}
	err := restoreStorage(server)
	if err != nil {
		panic(err)
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

func (s *httpServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	valType := chi.URLParam(r, "valType")
	name := chi.URLParam(r, "name")
	value := chi.URLParam(r, "value")
	if name == "" {
		http.NotFound(w, r)
		return
	}
	err := s.storage.Set(valType, name, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	answer := fmt.Sprintf(`Value %s type %s updated`, name, valType)
	_, err = w.Write([]byte(answer))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *httpServer) ValueHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain")
	valType := chi.URLParam(r, "valType")
	name := chi.URLParam(r, "name")
	if name == "" {
		http.NotFound(w, r)
		return
	}
	m, err := s.storage.Get(valType, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, err = w.Write([]byte(m.Value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *httpServer) ValuesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain")
	values, err := s.storage.Values()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var answer string
	for _, metric := range values {
		answer = answer + fmt.Sprintf("%s=%s\n", metric.ID, metric.Value)
	}
	_, err = w.Write([]byte(answer))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *httpServer) UpdateJSONHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	m := &metrx.Metrics{}
	err := decoder.Decode(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	ms, err := m.GetMetricsString()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = s.storage.Set(ms.MType, ms.ID, ms.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *httpServer) ValueJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	decoder := json.NewDecoder(r.Body)
	m := &metrx.Metrics{}
	err := decoder.Decode(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	ms, err := s.storage.Get(m.MType, m.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	m.ParseMetricsString(ms)
	encoder := json.NewEncoder(w)
	err = encoder.Encode(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *httpServer) InfoPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html")
	w.Write([]byte("<html><h1>Info page</h1></html>"))
}

func (s *httpServer) Shutdown() {
	saveStorage(s)
}

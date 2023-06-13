package server

import (
	"fmt"
	"net/http"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/go-chi/chi/v5"
)

type Server interface {
	Run() error
	MountHandlers()
}

type httpServer struct {
	Addr    string
	router  *chi.Mux
	storage metrx.MemStorage
}

func (s *httpServer) Run() error {
	return http.ListenAndServe(s.Addr, s.router)
}

func NewServer(addr string) Server {
	metricsStorage := metrx.NewMetricsStorage()
	server := &httpServer{
		addr,
		nil,
		metricsStorage,
	}
	return server
}

func (s *httpServer) MountHandlers() {

	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Post("/update/{valType}/{name}/{value}", s.UpdateHandler)
		r.Route("/value", func(r chi.Router) {
			r.Get("/", s.ValuesHandler)
			r.Get("/{valType}/{name}", s.ValueHandler)
		})
	})
	s.router = router
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
	valType := chi.URLParam(r, "valType")
	name := chi.URLParam(r, "name")
	if name == "" {
		http.NotFound(w, r)
		return
	}
	value, err := s.storage.Get(valType, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	_, err = w.Write([]byte(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *httpServer) ValuesHandler(w http.ResponseWriter, r *http.Request) {
	values, err := s.storage.Values()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var answer string
	for _, metric := range values {
		answer = answer + fmt.Sprintf("%s=%s\n", metric.Name, metric.Value)
	}
	_, err = w.Write([]byte(answer))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

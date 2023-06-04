package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Nexadis/metalert/internal/metrics"
)

type Server interface {
	Run() error
}

type httpServer struct {
	Addr    string
	handler http.Handler
	storage metrics.MemStorage
}

func (h *httpServer) Run() error {
	return http.ListenAndServe(h.Addr, h.handler)
}

func NewServer(addr string) Server {
	metricsStorage := metrics.NewMetricsStorage()
	mux := http.NewServeMux()
	server := &httpServer{
		addr,
		mux,
		metricsStorage,
	}
	mux.HandleFunc(`/update/`, server.UpdateHandler)
	return server
}

func (s *httpServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `Invalid method`, http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-type") != "text/plain" {
		http.Error(w, `Invalid Content-type`, http.StatusNotImplemented)
		return
	}
	q := r.URL.Path

	splitted := strings.SplitN(q, "/", 5)
	fmt.Printf("%v:%T\n", splitted[2], splitted[2])
	if len(splitted) < 5 {
		http.NotFound(w, r)
		return
	}
	_, valType, name, val := splitted[1], splitted[2], splitted[3], splitted[4]
	if name == "" {
		http.NotFound(w, r)
		return
	}
	err := s.storage.Set(valType, name, val)
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

	w.WriteHeader(http.StatusOK)
}

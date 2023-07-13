package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage/db"
	"github.com/Nexadis/metalert/internal/utils/logger"
	"github.com/go-chi/chi/v5"
)

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
	_, err = w.Write([]byte(m.GetValue()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *httpServer) ValuesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain")
	values, err := s.storage.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var answer string
	for _, metric := range values {
		answer = answer + fmt.Sprintf("%s=%s\n", metric.GetID(), metric.GetValue())
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
	m.ParseMetricsString(ms.(*metrx.MetricsString))
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

func (s *httpServer) DBPing(w http.ResponseWriter, r *http.Request) {
	db, ok := s.storage.(*db.DB)
	if ok {
		err := db.Ping()
		if err == nil {
			w.Write([]byte("DB is ok"))
			return
		}
		logger.Error(err)
	}
	http.Error(w, "DB is not connected", http.StatusInternalServerError)
}

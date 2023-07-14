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

func (s *httpServer) Update(w http.ResponseWriter, r *http.Request) {
	mtype := chi.URLParam(r, "mtype")
	id := chi.URLParam(r, "id")
	value := chi.URLParam(r, "value")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	err := s.storage.Set(r.Context(), mtype, id, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	answer := fmt.Sprintf(`Value %s type %s updated`, id, mtype)
	_, err = w.Write([]byte(answer))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *httpServer) Value(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain")
	mtype := chi.URLParam(r, "mtype")
	id := chi.URLParam(r, "id")
	logger.Info("Value Handler", mtype, id)
	if id == "" {
		http.NotFound(w, r)
		return
	}
	m, err := s.storage.Get(r.Context(), mtype, id)
	if err != nil {
		logger.Error(err)
		http.NotFound(w, r)
		return
	}
	_, err = w.Write([]byte(m.GetValue()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *httpServer) Values(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain")
	values, err := s.storage.GetAll(r.Context())
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

func (s *httpServer) UpdateJSON(w http.ResponseWriter, r *http.Request) {
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
	err = s.storage.Set(r.Context(), ms.MType, ms.ID, ms.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (s *httpServer) ValueJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	decoder := json.NewDecoder(r.Body)
	m := &metrx.Metrics{}
	err := decoder.Decode(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	ms, err := s.storage.Get(r.Context(), m.MType, m.ID)
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

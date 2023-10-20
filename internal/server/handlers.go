package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Nexadis/metalert/internal/models"
	"github.com/Nexadis/metalert/internal/storage/db"
	"github.com/Nexadis/metalert/internal/storage/mem"
	"github.com/Nexadis/metalert/internal/utils/logger"
)

// Update Обновление метрики с помощью REST
func (s *HTTPServer) Update(w http.ResponseWriter, r *http.Request) {
	mtype := chi.URLParam(r, "mtype")
	id := chi.URLParam(r, "id")
	value := chi.URLParam(r, "value")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	m, err := models.NewMetric(id, mtype, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = s.storage.Set(r.Context(), m)
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

// Value Получение метрики с помощью REST
func (s *HTTPServer) Value(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain")
	mtype := chi.URLParam(r, "mtype")
	id := chi.URLParam(r, "id")
	logger.Info("Value Handler", mtype, id)
	if id == "" {
		http.NotFound(w, r)
		return
	}
	m, err := s.storage.Get(r.Context(), mtype, id)
	if errors.Is(err, mem.ErrNotFound) {
		logger.Error(err)
		http.NotFound(w, r)
		return
	}
	val, err := m.GetValue()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = w.Write([]byte(val))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Values Возвращает все значения в текстовом формате
func (s *HTTPServer) Values(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain")
	values, err := s.storage.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var answer string
	for _, metric := range values {
		val, err := metric.GetValue()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		answer = answer + fmt.Sprintf("%s=%s\n", metric.ID, val)
	}
	_, err = w.Write([]byte(answer))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UpdateJSON Обработчик для записи одиночных метрик в JSON-формате
func (s *HTTPServer) UpdateJSON(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	m := &models.Metric{}
	err := decoder.Decode(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	err = s.storage.Set(r.Context(), *m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

// Updates Обработчик для записи списка метрик в JSON-формате
func (s *HTTPServer) Updates(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	metrics := make([]models.Metric, 0, 50)
	err := decoder.Decode(&metrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	logger.Info("Parse metrics in Updates handler")
	for _, m := range metrics {
		err = s.storage.Set(r.Context(), m)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

// ValueJSON Обработчик для получения одиночных метрик в JSON-формате
func (s *HTTPServer) ValueJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	decoder := json.NewDecoder(r.Body)
	m := &models.Metric{}
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
	encoder := json.NewEncoder(w)
	err = encoder.Encode(ms)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// InfoPage Главная страница - заглушка
func (s *HTTPServer) InfoPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html")
	_, err := w.Write([]byte("<html><h1>Info page</h1></html>"))
	if err != nil {
		logger.Error(err)
	}
}

// DBPing Проверяет состояние подключения к базе данных
func (s *HTTPServer) DBPing(w http.ResponseWriter, r *http.Request) {
	db, ok := s.storage.(*db.DB)
	if ok {
		err := db.Ping()
		if err == nil {
			_, err = w.Write([]byte("DB is ok"))
			if err != nil {
				logger.Error(err)
			}
			return
		}
		logger.Error(err)
	}
	http.Error(w, "DB is not connected", http.StatusInternalServerError)
}

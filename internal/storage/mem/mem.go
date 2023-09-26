// mem реализует хранилище с помощью map
package mem

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage"
)

type MetricsStorage interface {
	storage.Storage
	StateSaver
}

var _ MetricsStorage = &Storage{}

// Storage - Хранилище inmemory. Отдельно хранит Gauge и Counter метрики. Использует RWMutex Для доступа к элементам.
type Storage struct {
	Gauges   map[string]metrx.Gauge
	Counters map[string]metrx.Counter
	mutex    sync.RWMutex
}

// NewMetricsStorage Конструктор для Storage
func NewMetricsStorage() *Storage {
	ms := new(Storage)
	ms.Gauges = make(map[string]metrx.Gauge)
	ms.Counters = make(map[string]metrx.Counter)
	return ms
}

// Set Добавляет метрику
func (ms *Storage) Set(ctx context.Context, m metrx.Metrics) error {
	_, err := m.GetValue()
	if err != nil {
		return err
	}
	switch strings.ToLower(m.MType) {
	case metrx.CounterType:
		ms.mutex.Lock()
		ms.Counters[m.ID] += *m.Delta
		ms.mutex.Unlock()
		return nil
	case metrx.GaugeType:
		ms.mutex.Lock()
		ms.Gauges[m.ID] = *m.Value
		ms.mutex.Unlock()
		return nil
	}
	return fmt.Errorf("%v: %v", storage.ErrInvalidType, m)
}

// Get Получает метрику с типом mtype и именем id
func (ms *Storage) Get(ctx context.Context, mtype, id string) (metrx.Metrics, error) {
	switch strings.ToLower(mtype) {
	case metrx.CounterType:
		ms.mutex.RLock()
		value, ok := ms.Counters[id]
		ms.mutex.RUnlock()
		if !ok {
			return metrx.Metrics{}, storage.ErrNotFound
		}
		return metrx.NewMetrics(id, mtype, value.String())
	case metrx.GaugeType:
		ms.mutex.RLock()
		value, ok := ms.Gauges[id]
		ms.mutex.RUnlock()
		if !ok {
			return metrx.Metrics{}, storage.ErrNotFound
		}
		return metrx.NewMetrics(id, mtype, value.String())
	}

	return metrx.Metrics{}, storage.ErrInvalidType
}

// GetAll Получает все метрики из хранилища
func (ms *Storage) GetAll(ctx context.Context) ([]metrx.Metrics, error) {
	m := make([]metrx.Metrics, 0, len(ms.Gauges)+len(ms.Counters))
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	for name, value := range ms.Gauges {
		v := value
		m = append(m, metrx.Metrics{
			MType: metrx.GaugeType,
			ID:    name,
			Value: &v,
		})
	}
	for name, value := range ms.Counters {
		v := value
		m = append(m, metrx.Metrics{
			MType: metrx.CounterType,
			ID:    name,
			Delta: &v,
		})
	}
	return m, nil
}

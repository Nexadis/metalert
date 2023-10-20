// mem реализует хранилище с помощью map
package mem

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Nexadis/metalert/internal/models"
	"github.com/Nexadis/metalert/internal/storage"
)

// Storage - Хранилище inmemory. Отдельно хранит Gauge и Counter метрики. Использует RWMutex Для доступа к элементам.
type Storage struct {
	Gauges   map[string]models.Gauge
	Counters map[string]models.Counter
	mutex    sync.RWMutex
}

// NewMetricsStorage Конструктор для Storage
func NewMetricsStorage() *Storage {
	ms := new(Storage)
	ms.Gauges = make(map[string]models.Gauge)
	ms.Counters = make(map[string]models.Counter)
	return ms
}

// Set Добавляет метрику
func (ms *Storage) Set(ctx context.Context, m models.Metric) error {
	_, err := m.GetValue()
	if err != nil {
		return err
	}
	switch strings.ToLower(m.MType) {
	case models.CounterType:
		ms.mutex.Lock()
		ms.Counters[m.ID] += *m.Delta
		ms.mutex.Unlock()
		return nil
	case models.GaugeType:
		ms.mutex.Lock()
		ms.Gauges[m.ID] = *m.Value
		ms.mutex.Unlock()
		return nil
	}
	return fmt.Errorf("%v: %v", storage.ErrInvalidType, m)
}

// Get Получает метрику с типом mtype и именем id
func (ms *Storage) Get(ctx context.Context, mtype, id string) (models.Metric, error) {
	switch strings.ToLower(mtype) {
	case models.CounterType:
		ms.mutex.RLock()
		value, ok := ms.Counters[id]
		ms.mutex.RUnlock()
		if !ok {
			return models.Metric{}, storage.ErrNotFound
		}
		return models.NewMetric(id, mtype, value.String())
	case models.GaugeType:
		ms.mutex.RLock()
		value, ok := ms.Gauges[id]
		ms.mutex.RUnlock()
		if !ok {
			return models.Metric{}, storage.ErrNotFound
		}
		return models.NewMetric(id, mtype, value.String())
	}

	return models.Metric{}, storage.ErrInvalidType
}

// GetAll Получает все метрики из хранилища
func (ms *Storage) GetAll(ctx context.Context) ([]models.Metric, error) {
	m := make([]models.Metric, 0, len(ms.Gauges)+len(ms.Counters))
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	for name, value := range ms.Gauges {
		v := value
		m = append(m, models.Metric{
			MType: models.GaugeType,
			ID:    name,
			Value: &v,
		})
	}
	for name, value := range ms.Counters {
		v := value
		m = append(m, models.Metric{
			MType: models.CounterType,
			ID:    name,
			Delta: &v,
		})
	}
	return m, nil
}

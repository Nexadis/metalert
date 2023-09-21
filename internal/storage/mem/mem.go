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

type Storage struct {
	Gauges   map[string]metrx.Gauge
	Counters map[string]metrx.Counter
	mutex    sync.RWMutex
}

func NewMetricsStorage() *Storage {
	ms := new(Storage)
	ms.Gauges = make(map[string]metrx.Gauge)
	ms.Counters = make(map[string]metrx.Counter)
	return ms
}

func (ms *Storage) Set(ctx context.Context, m metrx.Metrics) error {
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

func (ms *Storage) Get(ctx context.Context, mtype, id string) (metrx.Metrics, error) {
	switch strings.ToLower(mtype) {
	case metrx.CounterType:
		ms.mutex.RLock()
		value, ok := ms.Counters[id]
		ms.mutex.RUnlock()
		if !ok {
			return metrx.Metrics{}, storage.ErrNotFound
		}
		return metrx.Metrics{
			ID:    id,
			MType: mtype,
			Delta: &value,
		}, nil
	case metrx.GaugeType:
		ms.mutex.RLock()
		value, ok := ms.Gauges[id]
		ms.mutex.RUnlock()
		if !ok {
			return metrx.Metrics{}, storage.ErrNotFound
		}
		return metrx.Metrics{
			ID:    id,
			MType: mtype,
			Value: &value,
		}, nil
	}

	return metrx.Metrics{}, storage.ErrInvalidType
}

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

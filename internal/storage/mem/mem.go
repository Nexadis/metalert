package mem

import (
	"context"
	"strings"
	"sync"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage"
)

type MetricsStorage interface {
	storage.Storage
	StateSaver
}

var _ storage.Storage = &Storage{}

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

func (ms *Storage) Set(ctx context.Context, mtype, id, value string) error {
	switch strings.ToLower(mtype) {
	case metrx.CounterType:
		val, err := metrx.ParseCounter(value)
		if err != nil {
			return err
		}
		ms.mutex.Lock()
		ms.Counters[id] += val
		ms.mutex.Unlock()
		return nil
	case metrx.GaugeType:
		val, err := metrx.ParseGauge(value)
		if err != nil {
			return err
		}
		ms.mutex.Lock()
		ms.Gauges[id] = val
		ms.mutex.Unlock()
		return nil
	}
	return storage.ErrInvalidType
}

func (ms *Storage) Get(ctx context.Context, mtype, id string) (*metrx.MetricsString, error) {
	switch strings.ToLower(mtype) {
	case metrx.CounterType:
		ms.mutex.RLock()
		value, ok := ms.Counters[id]
		ms.mutex.RUnlock()
		if !ok {
			return nil, storage.ErrNotFound
		}
		val := &metrx.MetricsString{
			ID:    id,
			MType: metrx.CounterType,
			Value: value.String(),
		}
		return val, nil
	case metrx.GaugeType:
		ms.mutex.RLock()
		value, ok := ms.Gauges[id]
		ms.mutex.RUnlock()
		if !ok {
			return nil, storage.ErrNotFound
		}
		val := &metrx.MetricsString{
			ID:    id,
			MType: metrx.GaugeType,
			Value: value.String(),
		}
		return val, nil
	}

	return nil, storage.ErrInvalidType
}

func (ms *Storage) GetAll(ctx context.Context) ([]*metrx.MetricsString, error) {
	m := make([]*metrx.MetricsString, 0, len(ms.Gauges)+len(ms.Counters))
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	for name, value := range ms.Gauges {
		val := value.String()
		m = append(m, &metrx.MetricsString{
			MType: metrx.GaugeType,
			ID:    name,
			Value: val,
		})
	}
	for name, value := range ms.Counters {
		val := value.String()
		m = append(m, &metrx.MetricsString{
			MType: metrx.CounterType,
			ID:    name,
			Value: val,
		})
	}
	return m, nil
}

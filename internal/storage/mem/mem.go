package mem

import (
	"context"
	"strings"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage"
)

type MetricsStorage interface {
	storage.Storage
	StateSaver
}

type Storage struct {
	Gauges   map[string]metrx.Gauge
	Counters map[string]metrx.Counter
}

func NewMetricsStorage() MetricsStorage {
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
		ms.Counters[id] += val
		return nil
	case metrx.GaugeType:
		val, err := metrx.ParseGauge(value)
		if err != nil {
			return err
		}
		ms.Gauges[id] = val
		return nil
	}
	return storage.ErrInvalidType
}

func (ms *Storage) Get(ctx context.Context, mtype, id string) (storage.ObjectGetter, error) {
	switch strings.ToLower(mtype) {
	case metrx.CounterType:
		value, ok := ms.Counters[id]
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
		value, ok := ms.Gauges[id]
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

func (ms *Storage) GetAll(ctx context.Context) ([]storage.ObjectGetter, error) {
	m := make([]storage.ObjectGetter, 0, len(ms.Gauges)+len(ms.Counters))
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

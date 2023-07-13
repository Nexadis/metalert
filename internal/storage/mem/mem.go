package mem

import (
	"errors"
	"strings"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage"
)

type StateSaver interface {
	Restore(FileStoragePath string, Restore bool) error
	Save(FileStoragePath string) error
	SaveTimer(FileStoragePath string, interval int64)
}

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

func (ms *Storage) Set(valType, name, value string) error {
	switch strings.ToLower(valType) {
	case metrx.CounterType:
		val, err := metrx.ParseCounter(value)
		if err != nil {
			return err
		}
		ms.Counters[name] += val
		return nil
	case metrx.GaugeType:
		val, err := metrx.ParseGauge(value)
		if err != nil {
			return err
		}
		ms.Gauges[name] = val
		return nil
	}
	return errors.New("invalid type")
}

func (ms *Storage) Get(valType, name string) (storage.ObjectGetter, error) {
	switch strings.ToLower(valType) {
	case metrx.CounterType:
		value, ok := ms.Counters[name]
		if !ok {
			return nil, errors.New("value not found")
		}
		val := &metrx.MetricsString{
			ID:    name,
			MType: metrx.CounterType,
			Value: value.String(),
		}
		return val, nil
	case metrx.GaugeType:
		value, ok := ms.Gauges[name]
		if !ok {
			return nil, errors.New("value not found")
		}
		val := &metrx.MetricsString{
			ID:    name,
			MType: metrx.GaugeType,
			Value: value.String(),
		}
		return val, nil
	}

	return nil, errors.New("invalid type")
}

func (ms *Storage) GetAll() ([]storage.ObjectGetter, error) {
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

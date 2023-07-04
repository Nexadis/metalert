package storage

import (
	"errors"
	"strings"

	"github.com/Nexadis/metalert/internal/metrx"
)

type Object interface {
	GetMType() string
	GetID() string
	GetValue() string
}

type Getter interface {
	Get(valType, name string) (Object, error)
	GetAll() ([]Object, error)
}

type Setter interface {
	Set(valType, name, value string) error
}

type MemStorage interface {
	Getter
	Setter
}

type MetricsStorage struct {
	Gauges   map[string]metrx.Gauge
	Counters map[string]metrx.Counter
}

func NewMetricsStorage() MemStorage {
	ms := new(MetricsStorage)
	ms.Gauges = make(map[string]metrx.Gauge)
	ms.Counters = make(map[string]metrx.Counter)
	return ms
}

func (ms *MetricsStorage) Set(valType, name, value string) error {
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

func (ms *MetricsStorage) Get(valType, name string) (Object, error) {
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

func (ms *MetricsStorage) GetAll() ([]Object, error) {
	m := make([]Object, 0, len(ms.Gauges)+len(ms.Counters))
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

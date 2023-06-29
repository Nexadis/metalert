package metrx

import (
	"errors"
	"strconv"
	"strings"
)

type MetricsGetter interface {
	Get(valType, name string) (string, error)
	Values() ([]Metrica, error)
}

type MetricsSetter interface {
	Set(valType, name, value string) error
}

type MemStorage interface {
	MetricsGetter
	MetricsSetter
}

type (
	Gauge   float64
	Counter int64
)

const (
	GaugeType   = `gauge`
	CounterType = `counter`
)

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func NewCounter(value string) (Counter, error) {
	val, err := strconv.Atoi(value)
	return Counter(val), err
}

func NewGauge(value string) (Gauge, error) {
	val, err := strconv.ParseFloat(value, 64)
	return Gauge(val), err
}

type Metrics struct {
	Gauges   map[string]Gauge
	Counters map[string]Counter
}

type Metrica struct {
	ValType string
	Name    string
	Value   string
}

func NewMetricsStorage() MemStorage {
	ms := new(Metrics)
	ms.Gauges = make(map[string]Gauge)
	ms.Counters = make(map[string]Counter)
	return ms
}

func (ms *Metrics) Set(valType, name, value string) error {
	switch strings.ToLower(valType) {
	case CounterType:
		val, err := NewCounter(value)
		if err != nil {
			return err
		}
		ms.Counters[name] += val
		return nil
	case GaugeType:
		val, err := NewGauge(value)
		if err != nil {
			return err
		}
		ms.Gauges[name] = val
		return nil
	}
	return errors.New("invalid type")
}

func (ms *Metrics) Get(valType, name string) (string, error) {
	switch strings.ToLower(valType) {
	case CounterType:
		value, ok := ms.Counters[name]
		if !ok {
			return "", errors.New("value not found")
		}
		val := value.String()
		return val, nil
	case GaugeType:
		value, ok := ms.Gauges[name]
		if !ok {
			return "", errors.New("value not found")
		}
		val := value.String()
		return val, nil
	}

	return "", errors.New("invalid type")
}

func (ms *Metrics) Values() ([]Metrica, error) {
	m := make([]Metrica, 0, len(ms.Gauges)+len(ms.Counters))
	for name, value := range ms.Gauges {
		val := value.String()
		m = append(m, Metrica{
			GaugeType,
			name,
			val,
		})
	}
	for name, value := range ms.Counters {
		val := value.String()
		m = append(m, Metrica{
			CounterType,
			name,
			val,
		})

	}
	return m, nil
}

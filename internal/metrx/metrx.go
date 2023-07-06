package metrx

import (
	"errors"
	"strconv"
	"strings"
)

type MetricsGetter interface {
	Get(valType, name string) (*MetricsString, error)
	Values() ([]*MetricsString, error)
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

var ErrorMetrics = errors.New("invalid metrics")

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func ParseCounter(value string) (Counter, error) {
	val, err := strconv.Atoi(value)
	return Counter(val), err
}

func ParseGauge(value string) (Gauge, error) {
	val, err := strconv.ParseFloat(value, 64)
	return Gauge(val), err
}

type MetricsStorage struct {
	Gauges   map[string]Gauge
	Counters map[string]Counter
}

type MetricsString struct {
	MType string
	ID    string
	Value string
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *Counter `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *Gauge   `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m *Metrics) ParseMetricsString(ms *MetricsString) error {
	m.ID = ms.ID
	m.MType = ms.MType
	switch ms.MType {
	case CounterType:
		v, err := ParseCounter(ms.Value)
		if err != nil {
			return err
		}
		m.Delta = &v
		m.Value = nil
	case GaugeType:
		v, err := ParseGauge(ms.Value)
		if err != nil {
			return err
		}
		m.Value = &v
		m.Delta = nil
	}
	return nil
}

func (m *Metrics) GetMetricsString() (*MetricsString, error) {
	ms := &MetricsString{
		ID:    m.ID,
		MType: m.MType,
	}
	switch m.MType {
	case CounterType:
		if m.Delta == nil {
			return nil, ErrorMetrics
		}
		ms.Value = m.Delta.String()
	case GaugeType:
		if m.Value == nil {
			return nil, ErrorMetrics
		}
		ms.Value = m.Value.String()
	}
	return ms, nil
}

func NewMetricsStorage() MemStorage {
	ms := new(MetricsStorage)
	ms.Gauges = make(map[string]Gauge)
	ms.Counters = make(map[string]Counter)
	return ms
}

func (ms *MetricsStorage) Set(valType, name, value string) error {
	switch strings.ToLower(valType) {
	case CounterType:
		val, err := ParseCounter(value)
		if err != nil {
			return err
		}
		ms.Counters[name] += val
		return nil
	case GaugeType:
		val, err := ParseGauge(value)
		if err != nil {
			return err
		}
		ms.Gauges[name] = val
		return nil
	}
	return errors.New("invalid type")
}

func (ms *MetricsStorage) Get(valType, name string) (*MetricsString, error) {
	switch strings.ToLower(valType) {
	case CounterType:
		value, ok := ms.Counters[name]
		if !ok {
			return nil, errors.New("value not found")
		}
		val := &MetricsString{
			ID:    name,
			MType: CounterType,
			Value: value.String(),
		}
		return val, nil
	case GaugeType:
		value, ok := ms.Gauges[name]
		if !ok {
			return nil, errors.New("value not found")
		}
		val := &MetricsString{
			ID:    name,
			MType: GaugeType,
			Value: value.String(),
		}
		return val, nil
	}

	return nil, errors.New("invalid type")
}

func (ms *MetricsStorage) Values() ([]*MetricsString, error) {
	m := make([]*MetricsString, 0, len(ms.Gauges)+len(ms.Counters))
	for name, value := range ms.Gauges {
		val := value.String()
		m = append(m, &MetricsString{
			GaugeType,
			name,
			val,
		})
	}
	for name, value := range ms.Counters {
		val := value.String()
		m = append(m, &MetricsString{
			CounterType,
			name,
			val,
		})

	}
	return m, nil
}

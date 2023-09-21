package metrx

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

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

type MetricsString struct {
	MType string
	ID    string
	Value string
}

func (m MetricsString) GetMType() string {
	return m.MType
}

func (m MetricsString) GetID() string {
	return m.ID
}

func (m MetricsString) GetValue() string {
	return m.Value
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *Counter `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *Gauge   `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func (m Metrics) CheckType() error {
	switch strings.ToLower(m.MType) {
	case GaugeType:
		if m.Value == nil {
			return fmt.Errorf("%v: %v", ErrorMetrics, m)
		}
	case CounterType:
		if m.Delta == nil {
			return fmt.Errorf("%v: %v", ErrorMetrics, m)
		}
	}
	return nil
}

func NewMetrics(id, mtype, value string) (Metrics, error) {
	m := Metrics{}
	ms := MetricsString{
		ID:    id,
		MType: mtype,
		Value: value,
	}
	err := m.ParseMetricsString(ms)
	return m, err
}

func (m *Metrics) ParseMetricsString(ms MetricsString) error {
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

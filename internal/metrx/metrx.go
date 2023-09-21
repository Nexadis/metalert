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

var (
	ErrorMetrics = errors.New("invalid metrics")
	ErrorType    = errors.New("invalid type")
)

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

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *Counter `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *Gauge   `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewMetrics(id, mtype, value string) (Metrics, error) {
	m := &Metrics{
		ID:    id,
		MType: strings.ToLower(mtype),
	}
	err := m.SetValue(value)
	return *m, err
}

func (m *Metrics) SetValue(value string) error {
	switch m.MType {
	case CounterType:
		v, err := ParseCounter(value)
		if err != nil {
			return fmt.Errorf("%v: %v", ErrorMetrics, err)
		}
		m.Delta = &v
		m.Value = nil
	case GaugeType:
		v, err := ParseGauge(value)
		if err != nil {
			return fmt.Errorf("%v: %v", ErrorMetrics, err)
		}
		m.Value = &v
		m.Delta = nil
	default:
		return fmt.Errorf("%v: %v", ErrorType, m)
	}
	return nil
}

func (m Metrics) GetValue() (string, error) {
	switch m.MType {
	case CounterType:
		if m.Delta == nil {
			return "", ErrorMetrics
		}
		return m.Delta.String(), nil
	case GaugeType:
		if m.Value == nil {
			return "", ErrorMetrics
		}
		return m.Value.String(), nil
	}
	return "", fmt.Errorf("%v: %v", ErrorType, m.MType)
}

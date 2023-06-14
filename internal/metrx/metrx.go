package metrx

import (
	"errors"
	"fmt"
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

type Gauge float64
type Counter int64

const GaugeType = `gauge`
const CounterType = `counter`

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
	switch {
	case strings.Compare(valType, CounterType) == 0:
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		ms.Counters[name] += Counter(val)
		return nil
	case strings.Compare(valType, GaugeType) == 0:
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		ms.Gauges[name] = Gauge(val)
		return nil
	}
	message := fmt.Sprintf("Ivalid type! Type is %s", strings.ToLower(valType))
	return errors.New(message)
}

func (ms *Metrics) Get(valType, name string) (string, error) {
	switch strings.ToLower(valType) {
	case CounterType:
		value, ok := ms.Counters[name]
		if !ok {
			return "", errors.New("value not found")
		}
		val := strconv.FormatInt(int64(value), 10)
		return val, nil
	case GaugeType:
		value, ok := ms.Gauges[name]
		if !ok {
			return "", errors.New("value not found")
		}
		val := strconv.FormatFloat(float64(value), 'f', -1, 64)
		return val, nil
	}

	return "", errors.New("invalid type")
}

func (ms *Metrics) Values() ([]Metrica, error) {
	m := make([]Metrica, 0, len(ms.Gauges)+len(ms.Counters))
	for name, value := range ms.Gauges {
		val := strconv.FormatFloat(float64(value), 'f', -1, 64)
		m = append(m, Metrica{
			GaugeType,
			name,
			val,
		})
	}
	for name, value := range ms.Counters {
		val := strconv.FormatInt(int64(value), 10)
		m = append(m, Metrica{
			CounterType,
			name,
			val,
		})

	}
	return m, nil
}

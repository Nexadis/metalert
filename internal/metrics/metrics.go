package metrics

import (
	"errors"
	"strconv"
	"strings"
)

type MetricsGetter interface {
	Get(name string) (string, error)
	Values() (map[string][]string, error)
}

type MetricsSetter interface {
	Set(name, valType, value string) error
}

type MetricsStorage interface {
	MetricsGetter
	MetricsSetter
}

type Gauge float64
type Counter int64

type Metrics struct {
	Gauges   map[string]Gauge
	Counters map[string]Counter
}

func (ms *Metrics) Set(name, valType, value string) error {
	switch strings.ToLower(valType) {
	case "counter":
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		ms.Counters[name] += Counter(val)
	case "gauge":
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		ms.Gauges[name] += Gauge(val)
	default:
		return errors.New("Invalid type")
	}

	return nil
}

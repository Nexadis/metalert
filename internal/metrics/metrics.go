package metrics

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type MetricsGetter interface {
	Get(valType, name string) (string, error)
	Values() (map[string]string, error)
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

type Metrics struct {
	Gauges   map[string]Gauge
	Counters map[string]Counter
}

func NewMetricsStorage() MemStorage {
	m := new(Metrics)
	m.Gauges = make(map[string]Gauge)
	m.Counters = make(map[string]Counter)
	return m
}

func (ms *Metrics) Set(valType, name, value string) error {
	switch {
	case strings.Compare(valType, `counter`) == 0:
		val, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		ms.Counters[name] += Counter(val)
		return nil
	case strings.Compare(valType, `gauge`) == 0:
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
	case "counter":
		val := strconv.FormatInt(int64(ms.Counters[name]), 10)
		return val, nil
	case "gauge":
		val := strconv.FormatFloat(float64(ms.Gauges[name]), 'b', 64, 64)
		return val, nil
	}

	return "", errors.New("invalid type")
}

func (ms *Metrics) Values() (map[string]string, error) {
	m := make(map[string]string, 100)
	return m, nil
}

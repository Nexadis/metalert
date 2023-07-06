package metrx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet(t *testing.T) {
	storage := MetricsStorage{
		Gauges: map[string]Gauge{
			"positive": Gauge(0),
			"small":    Gauge(0),
			"neg":      Gauge(0),
		},
		Counters: map[string]Counter{
			"positive": Counter(0),
			"big":      Counter(0),
		},
	}
	type setReq struct {
		valType string
		name    string
		value   string
	}
	testCasesCounters := []struct {
		name    string
		request setReq
		want    Counter
	}{
		{
			name: "Counter type, positive",
			request: setReq{
				valType: "counter",
				name:    "positive",
				value:   "2",
			},
			want: Counter(2),
		},
		{
			name: "Counter type, big num",
			request: setReq{
				valType: "counter",
				name:    "big",
				value:   "2985198054390",
			},
			want: Counter(2985198054390),
		},
	}

	testCasesGauges := []struct {
		name    string
		request setReq
		want    Gauge
	}{
		{
			name: "Gauge type, positive",
			request: setReq{
				valType: "gauge",
				name:    "positive",
				value:   "102391",
			},
			want: Gauge(102391),
		},
		{
			name: "Gauge type, very small",
			request: setReq{
				valType: "gauge",
				name:    "small",
				value:   "0.000000000001",
			},
			want: Gauge(0.000000000001),
		},
		{
			name: "Gauge type, negative",
			request: setReq{
				valType: "gauge",
				name:    "neg",
				value:   "-102391",
			},
			want: Gauge(-102391),
		},
	}
	for _, test := range testCasesCounters {
		t.Run(test.name, func(t *testing.T) {
			err := storage.Set(test.request.valType, test.request.name, test.request.value)
			assert.Equal(t, storage.Counters[test.request.name], test.want)
			assert.NoError(t, err)
		})
	}
	for _, test := range testCasesGauges {
		t.Run(test.name, func(t *testing.T) {
			err := storage.Set(test.request.valType, test.request.name, test.request.value)
			assert.Equal(t, storage.Gauges[test.request.name], test.want)
			assert.NoError(t, err)
		})
	}
}

func TestGet(t *testing.T) {
	storage := MetricsStorage{
		Gauges: map[string]Gauge{
			"positive": Gauge(102391),
			"small":    Gauge(0.000000000001),
			"neg":      Gauge(-102391),
		},
		Counters: map[string]Counter{
			"positive": Counter(2),
			"big":      Counter(2985198054390),
		},
	}
	type getReq struct {
		valType string
		name    string
	}
	testCases := []struct {
		name    string
		request getReq
		want    string
	}{
		{
			name: "Counter type, positive",
			request: getReq{
				valType: "counter",
				name:    "positive",
			},
			want: "2",
		},
		{
			name: "Counter type, big num",
			request: getReq{
				valType: "counter",
				name:    "big",
			},
			want: "2985198054390",
		},
		{
			name: "Gauge type, positive",
			request: getReq{
				valType: "gauge",
				name:    "positive",
			},
			want: "102391",
		},
		{
			name: "Gauge type, very small",
			request: getReq{
				valType: "gauge",
				name:    "small",
			},
			want: "0.000000000001",
		},
		{
			name: "Gauge type, negative",
			request: getReq{
				valType: "gauge",
				name:    "neg",
			},
			want: "-102391",
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := storage.Get(test.request.valType, test.request.name)
			assert.Equal(t, res.Value, test.want)
			assert.NoError(t, err)
		})
	}
}

func TestMetrics(t *testing.T) {
	tests := []struct {
		name      string
		metrics   MetricsString
		errStatus error
	}{
		{
			name: "Counter conversion",
			metrics: MetricsString{
				MType: CounterType,
				ID:    "name",
				Value: "1243123",
			},
			errStatus: nil,
		},

		{
			name: "Gauge conversion",
			metrics: MetricsString{
				MType: GaugeType,
				ID:    "name",
				Value: "124.857832974",
			},
			errStatus: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := Metrics{}
			m.ParseMetricsString(&test.metrics)
			newMetric, err := m.GetMetricsString()
			require.Equal(t, test.metrics, *newMetric)
			assert.NoError(t, err)
		})
	}
}

package mem

import (
	"context"
	"testing"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	storage := Storage{
		Gauges: map[string]metrx.Gauge{
			"positive": metrx.Gauge(0),
			"small":    metrx.Gauge(0),
			"neg":      metrx.Gauge(0),
		},
		Counters: map[string]metrx.Counter{
			"positive": metrx.Counter(0),
			"big":      metrx.Counter(0),
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
		want    metrx.Counter
	}{
		{
			name: "Counter type, positive",
			request: setReq{
				valType: "counter",
				name:    "positive",
				value:   "2",
			},
			want: metrx.Counter(2),
		},
		{
			name: "Counter type, big num",
			request: setReq{
				valType: "counter",
				name:    "big",
				value:   "2985198054390",
			},
			want: metrx.Counter(2985198054390),
		},
	}

	testCasesGauges := []struct {
		name    string
		request setReq
		want    metrx.Gauge
	}{
		{
			name: "Gauge type, positive",
			request: setReq{
				valType: "gauge",
				name:    "positive",
				value:   "102391",
			},
			want: metrx.Gauge(102391),
		},
		{
			name: "Gauge type, very small",
			request: setReq{
				valType: "gauge",
				name:    "small",
				value:   "0.000000000001",
			},
			want: metrx.Gauge(0.000000000001),
		},
		{
			name: "Gauge type, negative",
			request: setReq{
				valType: "gauge",
				name:    "neg",
				value:   "-102391",
			},
			want: metrx.Gauge(-102391),
		},
	}
	ctx := context.TODO()
	for _, test := range testCasesCounters {
		t.Run(test.name, func(t *testing.T) {
			err := storage.Set(ctx, test.request.valType, test.request.name, test.request.value)
			assert.Equal(t, storage.Counters[test.request.name], test.want)
			assert.NoError(t, err)
		})
	}
	for _, test := range testCasesGauges {
		t.Run(test.name, func(t *testing.T) {
			err := storage.Set(ctx, test.request.valType, test.request.name, test.request.value)
			assert.Equal(t, storage.Gauges[test.request.name], test.want)
			assert.NoError(t, err)
		})
	}
}

func TestGet(t *testing.T) {
	storage := Storage{
		Gauges: map[string]metrx.Gauge{
			"positive": metrx.Gauge(102391),
			"small":    metrx.Gauge(0.000000000001),
			"neg":      metrx.Gauge(-102391),
		},
		Counters: map[string]metrx.Counter{
			"positive": metrx.Counter(2),
			"big":      metrx.Counter(2985198054390),
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
	ctx := context.TODO()
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := storage.Get(ctx, test.request.valType, test.request.name)
			assert.Equal(t, res.GetValue(), test.want)
			assert.NoError(t, err)
		})
	}
}

func BenchmarkGet(b *testing.B) {
	storage := Storage{
		Gauges: map[string]metrx.Gauge{
			"positive": metrx.Gauge(102391),
			"small":    metrx.Gauge(0.000000000001),
			"neg":      metrx.Gauge(-102391),
		},
		Counters: map[string]metrx.Counter{
			"positive": metrx.Counter(2),
			"big":      metrx.Counter(2985198054390),
		},
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.Get(ctx, "counter", "big")
	}
}

func BenchmarkSet(b *testing.B) {
	storage := NewMetricsStorage()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.Set(ctx, "counter", "big", "48902183409")
	}
}

package mem

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage"
)

func TestSet(t *testing.T) {
	storage := NewMetricsStorage()
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
			m, err := metrx.NewMetrics(test.request.name, test.request.valType, test.request.value)
			assert.NoError(t, err)

			err = storage.Set(ctx, m)
			assert.Equal(t, storage.Counters[test.request.name], test.want)
			assert.NoError(t, err)
		})
	}
	for _, test := range testCasesGauges {
		t.Run(test.name, func(t *testing.T) {
			m, err := metrx.NewMetrics(test.request.name, test.request.valType, test.request.value)
			assert.NoError(t, err)
			err = storage.Set(ctx, m)
			assert.Equal(t, storage.Gauges[test.request.name], test.want)
			assert.NoError(t, err)
		})
	}
}

type getReq struct {
	valType string
	name    string
}
type getWant struct {
	value string
	err   error
}

func TestGet(t *testing.T) {
	s := NewMetricsStorage()
	s.Gauges = map[string]metrx.Gauge{
		"positive": metrx.Gauge(102391),
		"small":    metrx.Gauge(0.000000000001),
		"neg":      metrx.Gauge(-102391),
	}
	s.Counters = map[string]metrx.Counter{
		"positive": metrx.Counter(2),
		"big":      metrx.Counter(2985198054390),
	}

	testCases := []struct {
		name    string
		request getReq
		want    getWant
	}{
		{
			name: "Counter type, positive",
			request: getReq{
				valType: "counter",
				name:    "positive",
			},
			want: getWant{"2", nil},
		},
		{
			name: "Counter type, big num",
			request: getReq{
				valType: "counter",
				name:    "big",
			},
			want: getWant{"2985198054390", nil},
		},
		{
			name: "Gauge type, positive",
			request: getReq{
				valType: "gauge",
				name:    "positive",
			},
			want: getWant{"102391", nil},
		},
		{
			name: "Gauge type, very small",
			request: getReq{
				valType: "gauge",
				name:    "small",
			},
			want: getWant{"0.000000000001", nil},
		},
		{
			name: "Gauge type, negative",
			request: getReq{
				valType: "gauge",
				name:    "neg",
			},
			want: getWant{"-102391", nil},
		},
		{
			name: "Gauge type, not found",
			request: getReq{
				valType: metrx.GaugeType,
				name:    "notfound",
			},
			want: getWant{"", storage.ErrNotFound},
		},
		{
			name: "Ivalid type",
			request: getReq{
				valType: "invalid",
				name:    "invalid",
			},
			want: getWant{"", storage.ErrInvalidType},
		},
		{
			name: "Counter type, not found",
			request: getReq{
				valType: metrx.CounterType,
				name:    "notfound",
			},
			want: getWant{"", storage.ErrNotFound},
		},
	}
	ctx := context.TODO()
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			res, err := s.Get(ctx, test.request.valType, test.request.name)
			if err != nil {
				assert.Equal(t, test.want.err, err)
				return
			}
			ms, err := res.GetValue()
			assert.NoError(t, err)
			assert.Equal(t, ms, test.want.value)
		})
	}
}

func TestGetAll(t *testing.T) {
	s := NewMetricsStorage()
	s.Gauges = map[string]metrx.Gauge{
		"gauge1": metrx.Gauge(0.123),
		"gauge2": metrx.Gauge(0.533),
	}
	s.Counters = map[string]metrx.Counter{
		"c1": metrx.Counter(1),
		"c2": metrx.Counter(2),
	}
	ctx := context.Background()
	all, err := s.GetAll(ctx)
	assert.NoError(t, err)
	for _, m := range all {
		mtemp, err := s.Get(ctx, m.MType, m.ID)
		assert.NoError(t, err)
		assert.Equal(t, mtemp.Value, m.Value)
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

func randomMS(b *testing.B) metrx.Metrics {
	var mtype, val string
	value := rand.Int()
	if value%2 == 0 {
		mtype = "gauge"
		val = fmt.Sprintf("%d.%d", value, value)
	} else {
		mtype = "counter"
		val = fmt.Sprintf("%d", value)
	}
	name := val
	m, err := metrx.NewMetrics(name, mtype, val)
	assert.NoError(b, err)
	return m
}

func BenchmarkGetAll(b *testing.B) {
	storage := NewMetricsStorage()
	ctx := context.Background()
	for i := 0; i < 10000; i++ {
		m := randomMS(b)
		storage.Set(ctx, m)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.GetAll(ctx)
	}
}

func BenchmarkSet(b *testing.B) {
	storage := NewMetricsStorage()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		m := randomMS(b)
		b.StartTimer()
		storage.Set(ctx, m)
	}
}

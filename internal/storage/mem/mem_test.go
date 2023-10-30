package mem

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Nexadis/metalert/internal/models"
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
		want    models.Counter
	}{
		{
			name: "Counter type, positive",
			request: setReq{
				valType: "counter",
				name:    "positive",
				value:   "2",
			},
			want: models.Counter(2),
		},
		{
			name: "Counter type, big num",
			request: setReq{
				valType: "counter",
				name:    "big",
				value:   "2985198054390",
			},
			want: models.Counter(2985198054390),
		},
	}

	testCasesGauges := []struct {
		name    string
		request setReq
		want    models.Gauge
	}{
		{
			name: "Gauge type, positive",
			request: setReq{
				valType: "gauge",
				name:    "positive",
				value:   "102391",
			},
			want: models.Gauge(102391),
		},
		{
			name: "Gauge type, very small",
			request: setReq{
				valType: "gauge",
				name:    "small",
				value:   "0.000000000001",
			},
			want: models.Gauge(0.000000000001),
		},
		{
			name: "Gauge type, negative",
			request: setReq{
				valType: "gauge",
				name:    "neg",
				value:   "-102391",
			},
			want: models.Gauge(-102391),
		},
	}
	ctx := context.TODO()
	for _, test := range testCasesCounters {
		t.Run(test.name, func(t *testing.T) {
			m, err := models.NewMetric(test.request.name, test.request.valType, test.request.value)
			assert.NoError(t, err)

			err = storage.Set(ctx, m)
			assert.Equal(t, storage.Counters[test.request.name], test.want)
			assert.NoError(t, err)
		})
	}
	for _, test := range testCasesGauges {
		t.Run(test.name, func(t *testing.T) {
			m, err := models.NewMetric(test.request.name, test.request.valType, test.request.value)
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
	s.Gauges = map[string]models.Gauge{
		"positive": models.Gauge(102391),
		"small":    models.Gauge(0.000000000001),
		"neg":      models.Gauge(-102391),
	}
	s.Counters = map[string]models.Counter{
		"positive": models.Counter(2),
		"big":      models.Counter(2985198054390),
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
				valType: models.GaugeType,
				name:    "notfound",
			},
			want: getWant{"", ErrNotFound},
		},
		{
			name: "Ivalid type",
			request: getReq{
				valType: "invalid",
				name:    "invalid",
			},
			want: getWant{"", ErrInvalidType},
		},
		{
			name: "Counter type, not found",
			request: getReq{
				valType: models.CounterType,
				name:    "notfound",
			},
			want: getWant{"", ErrNotFound},
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
	s.Gauges = map[string]models.Gauge{
		"gauge1": models.Gauge(0.123),
		"gauge2": models.Gauge(0.533),
	}
	s.Counters = map[string]models.Counter{
		"c1": models.Counter(1),
		"c2": models.Counter(2),
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
		Gauges: map[string]models.Gauge{
			"positive": models.Gauge(102391),
			"small":    models.Gauge(0.000000000001),
			"neg":      models.Gauge(-102391),
		},
		Counters: map[string]models.Counter{
			"positive": models.Counter(2),
			"big":      models.Counter(2985198054390),
		},
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.Get(ctx, "counter", "big")
	}
}

func randomMS(b *testing.B) models.Metric {
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
	m, err := models.NewMetric(name, mtype, val)
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

package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	storage := Metrics{
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
			t.Log(res)
			assert.Equal(t, res, test.want)
			assert.NoError(t, err)
		})
	}
}

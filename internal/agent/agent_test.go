package agent

import (
	"net/http"
	"testing"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/stretchr/testify/assert"
)

var endpoint = "http://localhost:8080"

func TestPull(t *testing.T) {
	storage := metrx.NewMetricsStorage()
	ha := &httpAgent{
		listener:       endpoint,
		pullInterval:   0,
		reportInterval: 0,
		storage:        storage,
		client:         &http.Client{},
	}
	ha.Pull()
	type want struct {
		name    string
		valType string
		value   string
	}
	testsRuntime := []struct {
		name string
		want want
	}{
		{
			name: "Check StackSys",
			want: want{
				name:    "StackSys",
				valType: metrx.GaugeType,
				value:   "0",
			},
		},
		{
			name: "Check StackInuse",
			want: want{
				name:    "StackInuse",
				valType: metrx.GaugeType,
				value:   "0",
			},
		},
		{
			name: "Check PauseTotalNs",
			want: want{
				name:    "PauseTotalNs",
				valType: metrx.GaugeType,
				value:   "0",
			},
		},
		{
			name: "Check OtherSys",
			want: want{
				name:    "OtherSys",
				valType: metrx.GaugeType,
				value:   "0",
			},
		},
		{
			name: "Check TotalAlloc",
			want: want{
				name:    "TotalAlloc",
				valType: metrx.GaugeType,
				value:   "0",
			},
		},
		{
			name: "Check Sys",
			want: want{
				name:    "Sys",
				valType: metrx.GaugeType,
				value:   "0",
			},
		}}

	testsCounter := []struct {
		name string
		want want
	}{
		{
			name: "Check PollCount",
			want: want{
				name:    "PollCount",
				valType: metrx.CounterType,
				value:   "1",
			},
		},
	}
	for _, test := range testsRuntime {
		t.Run(test.name, func(t *testing.T) {
			value, err := ha.storage.Get(test.want.valType, test.want.name)
			assert.ErrorIs(t, nil, err)
			assert.NotEmpty(t, value)
		})
	}
	for _, test := range testsCounter {
		t.Run(test.name, func(t *testing.T) {
			value, err := ha.storage.Get(test.want.valType, test.want.name)
			assert.ErrorIs(t, nil, err)
			assert.NotEmpty(t, value)
			assert.Equal(t, test.want.value, value)
		})
	}

}

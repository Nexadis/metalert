package agent

import (
	"context"
	"testing"

	"github.com/Nexadis/metalert/internal/metrx"
	"github.com/Nexadis/metalert/internal/storage/mem"
	"github.com/stretchr/testify/assert"
)

var endpoint = "http://localhost:8080"

func TestPull(t *testing.T) {
	defineRuntimes()
	storage := mem.NewMetricsStorage()
	ha := &httpAgent{
		listener:       endpoint,
		pullInterval:   0,
		reportInterval: 0,
		storage:        storage,
		client:         nil,
	}
	err := ha.Pull()
	if err != nil {
		t.Error(err)
	}
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
		},
	}

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
	ctx := context.TODO()
	for _, test := range testsRuntime {
		t.Run(test.name, func(t *testing.T) {
			value, err := ha.storage.Get(ctx, test.want.valType, test.want.name)
			assert.ErrorIs(t, nil, err)
			assert.NotEmpty(t, value)
		})
	}
	for _, test := range testsCounter {
		t.Run(test.name, func(t *testing.T) {
			value, err := ha.storage.Get(ctx, test.want.valType, test.want.name)
			assert.ErrorIs(t, nil, err)
			assert.NotEmpty(t, value)
			assert.Equal(t, test.want.value, value.GetValue())
		})
	}
}

type testClient struct {
	path    string
	valType string
	name    string
	value   string
}

func (c *testClient) Post(path, valType, name, value string) error {
	c.path = path
	c.valType = valType
	c.name = name
	c.value = value
	return nil
}

func (c *testClient) PostObj(path string, obj interface{}) error {
	return nil
}

func TestReport(t *testing.T) {
	type want struct {
		name    string
		valType string
		value   string
	}
	tests := []struct {
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
		},
		{
			name: "Check PollCount",
			want: want{
				name:    "PollCount",
				valType: metrx.CounterType,
				value:   "1",
			},
		},
	}
	ctx := context.TODO()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testClient := &testClient{}
			storage := mem.NewMetricsStorage()
			ha := &httpAgent{
				listener:       endpoint,
				pullInterval:   0,
				reportInterval: 0,
				storage:        storage,
				client:         testClient,
			}
			err := ha.storage.Set(ctx, test.want.valType, test.want.name, test.want.value)
			assert.NoError(t, err)
			err = ha.Report()
			assert.NoError(t, err)
			assert.Equal(t, test.want.name, testClient.name)
			assert.Equal(t, test.want.valType, testClient.valType)
			assert.Equal(t, test.want.value, testClient.value)
		})
	}
}

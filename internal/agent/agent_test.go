package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Nexadis/metalert/internal/metrx"
)

var endpoint = "http://localhost:8080"

func TestPull(t *testing.T) {
	defineRuntimes()
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
	config := &Config{
		Address:        endpoint,
		PollInterval:   0,
		ReportInterval: 0,
	}
	ha := &httpAgent{
		config:  config,
		client:  nil,
		counter: 0,
	}
	mchan := make(chan *metrx.Metrics, 100)
	ha.Pull(context.Background(), mchan)
	close(mchan)
	metrics := make(map[string]metrx.Metrics, 100)
	for m := range mchan {
		metrics[m.ID] = *m
	}
	for _, test := range testsRuntime {
		t.Run(test.name, func(t *testing.T) {
			_, ok := metrics[test.want.name]
			assert.True(t, ok)
		})
	}
	for _, test := range testsCounter {
		t.Run(test.name, func(t *testing.T) {
			value, ok := metrics[test.want.name]
			assert.True(t, ok)
			assert.NotEmpty(t, value)
			v, err := value.GetValue()
			assert.NoError(t, err)
			assert.Equal(t, test.want.value, v)
		})
	}
}

type testClient struct {
	path    string
	valType string
	name    string
	value   string
}

func (c *testClient) Post(ctx context.Context, path, valType, name, value string) error {
	c.path = path
	c.valType = valType
	c.name = name
	c.value = value
	return nil
}

func (c *testClient) PostObj(ctx context.Context, path string, obj interface{}) error {
	return nil
}

func TestReport(t *testing.T) {
	t.Log("Run goroutine report")
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
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testClient := &testClient{}
			config := &Config{
				Address:        endpoint,
				PollInterval:   0,
				ReportInterval: 0,
			}
			ha := &httpAgent{
				config: config,
				client: testClient,
			}
			mchan := make(chan *metrx.Metrics, 1)
			errs := make(chan error, 1)
			ctx := context.Background()
			m, err := metrx.NewMetrics(test.want.name, test.want.valType, test.want.value)
			assert.NoError(t, err)
			mchan <- &m
			close(mchan)
			ha.Report(ctx, mchan, errs)
			ctx.Done()
			assert.Equal(t, test.want.name, testClient.name)
			assert.Equal(t, test.want.valType, testClient.valType)
			assert.Equal(t, test.want.value, testClient.value)
			close(errs)
		})
	}
}

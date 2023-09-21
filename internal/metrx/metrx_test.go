package metrx

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConversions(t *testing.T) {
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
			m.ParseMetricsString(test.metrics)
			newMetric, err := m.GetMetricsString()
			require.Equal(t, test.metrics, *newMetric)
			assert.NoError(t, err)
		})
	}
}

func randomMS() MetricsString {
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
	return MetricsString{
		ID:    name,
		MType: mtype,
		Value: val,
	}
}

func BenchmarkConversion(b *testing.B) {
	m := Metrics{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ms := randomMS()
		b.StartTimer()
		m.ParseMetricsString(ms)
	}
}

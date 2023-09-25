package metrx

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testMetric struct {
	ID    string
	MType string
	Value string
}

func TestParsing(t *testing.T) {
	tests := []struct {
		name      string
		metrics   testMetric
		errStatus error
	}{
		{
			name: "Counter conversion",
			metrics: testMetric{
				MType: CounterType,
				ID:    "name",
				Value: "1243123",
			},
			errStatus: nil,
		},

		{
			name: "Gauge conversion",
			metrics: testMetric{
				MType: GaugeType,
				ID:    "name",
				Value: "124.857832974",
			},
			errStatus: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			newMetric, err := NewMetrics(test.metrics.ID, test.metrics.MType, test.metrics.Value)
			assert.NoError(t, err)
			require.Equal(t, test.metrics.ID, newMetric.ID)
			require.Equal(t, test.metrics.MType, newMetric.MType)
			val, err := newMetric.GetValue()
			assert.NoError(t, err)
			require.Equal(t, test.metrics.Value, val)
		})
	}
}

func randomMS(b *testing.B) Metrics {
	b.StopTimer()
	var val string
	value := rand.Int()
	if value%2 == 0 {
		val = fmt.Sprintf("%d.%d", value, value)
	} else {
		val = fmt.Sprintf("%d", value)
	}
	b.StartTimer()
	m, _ := NewMetrics(val, GaugeType, val)
	return m
}

func BenchmarkConversion(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		randomMS(b)
	}
}

func ExampleMetrics() {
	m, err := NewMetrics("test", GaugeType, "1.1")
	if err != nil {
		panic(err)
	}
	v, err := m.GetValue()
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
	err = m.SetValue("2.2")
	if err != nil {
		panic(err)
	}
	v, err = m.GetValue()
	if err != nil {
		panic(err)
	}
	fmt.Println(v)

	m, err = NewMetrics("test", CounterType, "434")
	if err != nil {
		panic(err)
	}
	v, err = m.GetValue()
	if err != nil {
		panic(err)
	}
	fmt.Println(v)

	err = m.SetValue("567")
	if err != nil {
		panic(err)
	}
	v, err = m.GetValue()
	if err != nil {
		panic(err)
	}
	fmt.Println(v)

	// Output:
	// 1.1
	// 2.2
	// 434
	// 567
}

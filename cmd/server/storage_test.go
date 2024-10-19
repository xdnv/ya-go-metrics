package main

import (
	"fmt"
	"strconv"
	"testing"

	"internal/app"
	. "internal/ports/storage"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	type want struct {
		Value         interface{}
		ValueSequence []interface{}
	}
	tests := []struct {
		name           string
		metricType     string
		metricName     string
		initialValue   interface{}
		updateSequence []interface{}
		want           want
	}{
		{
			name:           "001 Gauge test",
			metricType:     "gauge",
			metricName:     "Type1G",
			initialValue:   float64(0),
			updateSequence: []interface{}{float64(0.011), float64(0.012)},
			want: want{
				Value:         float64(0.012),
				ValueSequence: []interface{}{float64(0.011), float64(0.012)},
			},
		},
		{
			name:           "002 Counter test",
			metricType:     "counter",
			metricName:     "Type2C",
			initialValue:   int64(0.0),
			updateSequence: []interface{}{int64(50), int64(60)},
			want: want{
				Value:         int64(110),
				ValueSequence: []interface{}{int64(50), int64(110)},
			},
		},
	}

	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		var testStorage = NewMemStorage()
	// 		var tm Metric

	// 		switch tt.metricType {
	// 		case "gauge":
	// 			tm = &Gauge{Value: tt.initialValue.(float64)}
	// 		case "counter":
	// 			tm = &Counter{Value: tt.initialValue.(int64)}
	// 		default:
	// 			assert.True(t, true, fmt.Sprintf("Unsupported metric type: %s", tt.metricType))
	// 		}

	// 		for _, v := range tt.updateSequence {
	// 			tm.UpdateValue(v)
	// 		}

	// 		testStorage.Metrics[tt.metricName] = tm

	// 		assert.Equal(t, tt.want.Value, GetMetricValue(testStorage.Metrics[tt.metricName]))
	// 	})
	// }

	var testSc = new(app.ServerConfig)
	testSc.StorageMode = app.Memory

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var tstor = NewUniStorage(testSc)
			var tm Metric

			switch tt.metricType {
			case "gauge":
				tm = &Gauge{Value: tt.initialValue.(float64)}
			case "counter":
				tm = &Counter{Value: tt.initialValue.(int64)}
			default:
				assert.True(t, true, fmt.Sprintf("Unsupported metric type: %s", tt.metricType))
			}

			for i, v := range tt.updateSequence {
				tm.UpdateValue(v)
				tstor.SetMetric(tt.metricName, tm)

				//read back temporary metric state
				metric, _ := tstor.GetMetric(tt.metricName)
				assert.Equal(t, tt.want.ValueSequence[i], metric.GetValue())
			}

			//read back final metric value
			metric, _ := tstor.GetMetric(tt.metricName)
			assert.Equal(t, tt.want.Value, metric.GetValue())
		})
	}
}

// MemStorage performance benchmark
func BenchmarkMemStorage(b *testing.B) {
	var testSc = new(app.ServerConfig)
	testSc.StorageMode = app.Memory
	var tstor = NewUniStorage(testSc)

	b.Run("gauges_update", func(b *testing.B) {
		var v = 0.2
		n := "Gauge"
		tm := &Gauge{Value: v}
		for i := 0; i < b.N; i++ {
			tm.UpdateValue(v)
		}
		tstor.SetMetric(n, tm)
	})
	b.Run("counters_update", func(b *testing.B) {
		var v int64 = 1
		n := "Counter"
		tm := &Counter{Value: v}
		for i := 0; i < b.N; i++ {
			tm.UpdateValue(v)
		}
		tstor.SetMetric(n, tm)
	})
	b.Run("gauges_add", func(b *testing.B) {
		var v = 0.2
		n := "Gauge"
		for i := 0; i < b.N; i++ {
			tm := &Gauge{Value: v}
			tstor.SetMetric(n+strconv.Itoa(i), tm)
		}
	})
	b.Run("counters_add", func(b *testing.B) {
		var v int64 = 1
		n := "Counter"
		for i := 0; i < b.N; i++ {
			tm := &Counter{Value: v}
			tstor.SetMetric(n+strconv.Itoa(i), tm)
		}
	})
}

package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	// // это пойдёт в тесты
	// g := Gauge{Value: 0.0} //new(Gauge)
	// g.UpdateValue(0.011)
	// g.UpdateValue(0.012)

	// c := Counter{Value: 0} //new(Counter)
	// c.UpdateValue(50)
	// c.UpdateValue(60)

	// storage.Metrics["Type1G"] = g // append
	// storage.Metrics["Type2C"] = c // append

	// fmt.Printf("Gauge Metric: %v\n", GetMetricValue(storage.Metrics["Type1G"]))
	// fmt.Printf("Counter Metric: %v\n", GetMetricValue(storage.Metrics["Type2C"]))

	type want struct {
		Value interface{}
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
				Value: float64(0.012),
			},
		},
		{
			name:           "002 Counter test",
			metricType:     "counter",
			metricName:     "Type2C",
			initialValue:   int64(0.0),
			updateSequence: []interface{}{int64(50), int64(60)},
			want: want{
				Value: int64(110),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testStorage = NewMemStorage()
			var tm Metric

			switch tt.metricType {
			case "gauge":
				tm = &Gauge{Value: tt.initialValue.(float64)} //new(Gauge)
			case "counter":
				tm = &Counter{Value: tt.initialValue.(int64)} //new(Counter)
			default:
				assert.True(t, true, fmt.Sprintf("Unsupported metric type: %s", tt.metricType)) //throw error
			}

			for _, v := range tt.updateSequence {
				tm.UpdateValue(v)
			}

			testStorage.Metrics[tt.metricName] = tm // append

			assert.Equal(t, tt.want.Value, GetMetricValue(testStorage.Metrics[tt.metricName]))
		})
	}
}

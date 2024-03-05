package main

import (
	"fmt"
	"testing"

	. "internal/ports/storage"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
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
				tm = &Gauge{Value: tt.initialValue.(float64)}
			case "counter":
				tm = &Counter{Value: tt.initialValue.(int64)}
			default:
				assert.True(t, true, fmt.Sprintf("Unsupported metric type: %s", tt.metricType))
			}

			for _, v := range tt.updateSequence {
				tm.UpdateValue(v)
			}

			testStorage.Metrics[tt.metricName] = tm

			assert.Equal(t, tt.want.Value, GetMetricValue(testStorage.Metrics[tt.metricName]))
		})
	}
}

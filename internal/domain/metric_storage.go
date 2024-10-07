package domain

import "sync"

// internal metric storage object
type MetricStorage struct {
	sync.RWMutex `json:"-"`
	Gauges       map[string]float64
	Counters     map[string]int64
}

// metric storage fabric
func NewMetricStorage() *MetricStorage {
	var ms MetricStorage
	ms.Gauges = make(map[string]float64)
	ms.Counters = make(map[string]int64)
	return &ms
}

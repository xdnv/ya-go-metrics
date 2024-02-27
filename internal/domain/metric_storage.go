package domain

import "sync"

type MetricStorage struct {
	sync.RWMutex `json:"-"`
	Gauges       map[string]float64
	Counters     map[string]int64
}

func NewMetricStorage() *MetricStorage {
	var ms MetricStorage
	ms.Gauges = make(map[string]float64)
	ms.Counters = make(map[string]int64)
	return &ms
}

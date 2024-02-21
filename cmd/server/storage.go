package main

import (
	"strconv"
	//"sync"
)

// metric structure
type MetricMap map[string]Metric

type Metric interface {
	GetValue() interface{}
	//UpdateValue(interface{}) interface{}
	//UpdateValueS(string) error
}

func GetMetricValue(t Metric) interface{} {
	return t.GetValue()
}

// main metric storage
type MemStorage struct {
	//mu      sync.Mutex //TODO: https://go.dev/tour/concurrency/9
	Metrics MetricMap
}

// Gauge - float64
type Gauge struct {
	Value float64
}

func (t Gauge) GetValue() interface{} {
	return t.Value
}

func (t *Gauge) UpdateValue(metricValue float64) {
	//REPLACE value
	t.Value = metricValue
}

func (t *Gauge) UpdateValueS(metricValueS string) error {
	val, err := strconv.ParseFloat(metricValueS, 64)
	if err != nil {
		return err
	}

	t.Value = val
	//t.UpdateValue(val)
	return nil
}

// Counter - int64
type Counter struct {
	Value int64
}

func (t Counter) GetValue() interface{} {
	return t.Value
}

func (t *Counter) UpdateValue(metricValue int64) {
	//INCREMENT value
	t.Value += metricValue
}

func (t *Counter) UpdateValueS(metricValueS string) error {
	val, err := strconv.ParseInt(metricValueS, 10, 64)
	if err != nil {
		return err
	}

	t.Value += val
	//t.UpdateValue(val)
	return nil
}

// init metric storage
func NewMemStorage() MemStorage {
	return MemStorage{
		Metrics: make(MetricMap),
	}
}

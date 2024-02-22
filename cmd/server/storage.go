package main

import (
	"fmt"
	"strconv"
)

// metric structure
type MetricMap map[string]Metric

type Metric interface {
	GetType() string
	GetValue() interface{}
	UpdateValue(interface{})
	UpdateValueS(string) error
}

func GetMetricValue(t Metric) interface{} {
	return t.GetValue()
}

// main metric storage
type MemStorage struct {
	Metrics MetricMap
}

func (t MemStorage) UpdateMetricS(mType string, mName string, mValue string) error {
	var val Metric
	var ok bool

	switch mType {
	case "gauge":
		val, ok = t.Metrics[mName].(*Gauge)
		if !ok {
			val = &Gauge{}
			storage.Metrics[mName] = val.(*Gauge)
		}
	case "counter":
		val, ok = t.Metrics[mName].(*Counter)
		if !ok {
			val = &Counter{}
			storage.Metrics[mName] = val.(*Counter)
		}
	default:
		return fmt.Errorf("unexpected metric type: %s", mType)
	}

	err := val.UpdateValueS(mValue)
	if err != nil {
		return err
	}

	return nil
}

// Gauge - float64
type Gauge struct {
	Value float64
}

func (t Gauge) GetType() string {
	return "gauge"
}

func (t Gauge) GetValue() interface{} {
	return t.Value
}

func (t *Gauge) UpdateValue(metricValue interface{}) {
	//REPLACE value
	t.Value = metricValue.(float64)
}

func (t *Gauge) UpdateValueS(metricValueS string) error {
	val, err := strconv.ParseFloat(metricValueS, 64)
	if err != nil {
		return err
	}

	t.UpdateValue(val)

	return nil
}

// Counter - int64
type Counter struct {
	Value int64
}

func (t Counter) GetType() string {
	return "counter"
}

func (t Counter) GetValue() interface{} {
	return t.Value
}

func (t *Counter) UpdateValue(metricValue interface{}) {
	//INCREMENT value
	t.Value += metricValue.(int64)
}

func (t *Counter) UpdateValueS(metricValueS string) error {
	val, err := strconv.ParseInt(metricValueS, 10, 64)
	if err != nil {
		return err
	}

	t.UpdateValue(val)

	return nil
}

// init metric storage
func NewMemStorage() MemStorage {
	return MemStorage{
		Metrics: make(MetricMap),
	}
}

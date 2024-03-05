package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// serializable storage class for JSON exchange
type SerializableMetric struct {
	ID         string  `json:"id"`
	MType      string  `json:"type"`
	IntValue   int64   `json:"counter_value,omitempty"`
	FloatValue float64 `json:"gauge_value,omitempty"`
}

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
			t.Metrics[mName] = val.(*Gauge)
		}
	case "counter":
		val, ok = t.Metrics[mName].(*Counter)
		if !ok {
			val = &Counter{}
			t.Metrics[mName] = val.(*Counter)
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

func (t MemStorage) GetSerializableMetric(name string) (*SerializableMetric, error) {

	sm := new(SerializableMetric)
	sm.ID = name

	metric, ok := t.Metrics[name]

	if !ok {
		return nil, fmt.Errorf("metric not found: %s", name)
	}

	sm.MType = metric.GetType()

	switch sm.MType {
	case "gauge":
		sm.FloatValue = metric.GetValue().(float64)
	case "counter":
		sm.IntValue = metric.GetValue().(int64)
	default:
		return nil, fmt.Errorf("unexpected metric type: %s", sm.MType)
	}

	return sm, nil
}

func (t MemStorage) GetSerializableStorage() ([]SerializableMetric, error) {

	sma := []SerializableMetric{}

	for k := range t.Metrics {
		sm, err := t.GetSerializableMetric(k)
		if err != nil {
			return nil, err
		}
		sma = append(sma, *sm)
	}

	return sma, nil
}

// Save server state
func (t MemStorage) SaveState(fname string) error {

	sma, err := t.GetSerializableStorage()
	if err != nil {
		return err
	}

	// serialize to JSON
	data, err := json.Marshal(sma)
	if err != nil {
		return err
	}
	// save serialized data to file
	return os.WriteFile(fname, data, 0666)
}

// Load server state
func (t MemStorage) LoadState(fname string) error {

	sma := []SerializableMetric{}

	data, err := os.ReadFile(fname)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &sma); err != nil {
		return err
	}

	//clear all existing metrics
	for k := range t.Metrics {
		delete(t.Metrics, k)
	}

	for _, sm := range sma {
		switch sm.MType {
		case "gauge":
			t.Metrics[sm.ID] = &Gauge{Value: sm.FloatValue}
		case "counter":
			t.Metrics[sm.ID] = &Counter{Value: sm.IntValue}
		default:
			return fmt.Errorf("unknown metric type: %s", sm.MType)
		}
	}

	return nil
}

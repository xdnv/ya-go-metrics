package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"internal/domain"
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
	var metric Metric
	var ok bool

	switch mType {
	case "gauge":
		metric, ok = t.Metrics[mName].(*Gauge)
		if !ok {
			metric = &Gauge{}
			t.Metrics[mName] = metric.(*Gauge)
		}
	case "counter":
		metric, ok = t.Metrics[mName].(*Counter)
		if !ok {
			metric = &Counter{}
			t.Metrics[mName] = metric.(*Counter)
		}
	default:
		return fmt.Errorf("unexpected metric type: %s", mType)
	}

	err := metric.UpdateValueS(mValue)
	if err != nil {
		return err
	}

	return nil
}

func (t MemStorage) BatchUpdateMetrics(m *[]domain.Metrics, errs *[]error) *[]domain.Metrics {

	for _, v := range *m {
		mr := new(domain.MetricRequest)
		mr.Type = v.MType
		mr.Name = v.ID

		switch mr.Type {
		case "gauge":
			mr.Value = fmt.Sprint(*v.Value)
		case "counter":
			mr.Value = fmt.Sprint(*v.Delta)
		default:
			err := fmt.Errorf("ERROR: unsupported metric type %s", mr.Type)
			*errs = append(*errs, err)
			//http.Error(w, fmt.Sprintf("ERROR: unsupported metric type %s", mr.Type), http.StatusNotFound)
			continue
		}

		err := t.UpdateMetricS(mr.Type, mr.Name, mr.Value)
		if err != nil {
			fmt.Printf("UPDATE ERROR: %s", err.Error())
			*errs = append(*errs, err)
			continue
		}
	}

	return m
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
func NewMemStorage() *MemStorage {
	return &MemStorage{
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

// assign metric object to certain name. use with caution, TODO: replace with safer API
func (t MemStorage) SetMetric(name string, metric Metric) {
	t.Metrics[name] = metric
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

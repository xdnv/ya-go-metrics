package main

import (
	"strconv"
	//"sync"
)

//(1.1) принимает и хранит произвольные метрики двух типов:
//	- Тип gauge, float64 — новое значение должно замещать предыдущее.
//	- Тип counter, int64 — новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу.

// структура метрик
type MetricMap map[string]interface{}

// блок метрик. позже будет вынесен в отдельный модуль
type Metric interface {
	GetValue() interface{}
	//UpdateValue(interface{}) interface{}
	//UpdateValueS(string) error
}

func GetMetricValue(t Metric) interface{} {
	return t.GetValue()
}

// основное хранилище метрик
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
	//ЗАМЕНА значения
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
	//ИНКРЕМЕНТ значения
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

// структура хранения метрик
//var MetricValues = make(MetricMap)

// структура хранения метрик
func InitStorage() MemStorage {
	return MemStorage{
		Metrics: make(MetricMap),
	}
}

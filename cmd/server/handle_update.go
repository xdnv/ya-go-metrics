package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// iter7 storage class for JSON exchange
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// HTTP update processing
func updateMetricV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mr := new(MetricRequest)
	mr.Type = chi.URLParam(r, "type")
	mr.Name = chi.URLParam(r, "name")
	mr.Value = chi.URLParam(r, "value")

	err := storage.UpdateMetricS(mr.Type, mr.Name, mr.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//w.WriteHeader(http.StatusOK)
}

// HTTP update processing
func updateMetricV2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		fmt.Printf("TRACE ERROR: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mr := new(MetricRequest)
	mr.Type = m.MType
	mr.Name = m.ID

	switch mr.Type {
	case "gauge":
		mr.Value = fmt.Sprint(*m.Value)
	case "counter":
		mr.Value = fmt.Sprint(*m.Delta)
	default:
		http.Error(w, fmt.Sprintf("ERROR: unsupported metric type %s", mr.Type), http.StatusBadRequest)
	}

	err := storage.UpdateMetricS(mr.Type, mr.Name, mr.Value)
	if err != nil {
		fmt.Printf("TRACE ERROR: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metric := storage.Metrics[mr.Name]

	switch mr.Type {
	case "gauge":
		metricValue := metric.GetValue().(float64)
		m.Value = &metricValue
	case "counter":
		metricValue := metric.GetValue().(int64)
		m.Delta = &metricValue
	}

	resp, err := json.Marshal(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

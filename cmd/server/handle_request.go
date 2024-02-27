package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// HTTP request processing
func requestMetricV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mr := new(MetricRequest)
	mr.Type = chi.URLParam(r, "type")
	mr.Name = chi.URLParam(r, "name")

	//type validation
	switch mr.Type {
	case "gauge":
	case "counter":
	default:
		http.Error(w, fmt.Sprintf("unexpected metric type: %s", mr.Type), http.StatusBadRequest)
		return
	}

	val, ok := storage.Metrics[mr.Name]
	if !ok {
		http.Error(w, "Metric not found: "+mr.Name, http.StatusNotFound)
		return
	}

	//write metric value in plaintext
	body := fmt.Sprintf("%v", val.GetValue())
	_, _ = w.Write([]byte(body))

	//w.WriteHeader(http.StatusOK)
}

// HTTP request processing
func requestMetricV2(w http.ResponseWriter, r *http.Request) {
	mr := new(MetricRequest)
	mr.Type = chi.URLParam(r, "type")
	mr.Name = chi.URLParam(r, "name")

	//type validation
	switch mr.Type {
	case "gauge":
	case "counter":
	default:
		http.Error(w, fmt.Sprintf("unexpected metric type: %s", mr.Type), http.StatusBadRequest)
		return
	}

	metric, ok := storage.Metrics[mr.Name]
	if !ok {
		http.Error(w, "Metric not found: "+mr.Name, http.StatusNotFound)
		return
	}

	var m Metrics
	m.MType = mr.Type
	m.ID = mr.Name

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

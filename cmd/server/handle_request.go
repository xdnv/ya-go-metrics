package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"internal/adapters/logger"
	"internal/domain"

	"github.com/go-chi/chi/v5"
)

// HTTP single metric request v1 processing
func handleRequestMetricV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mr := new(domain.MetricRequest)
	mr.Type = chi.URLParam(r, "type")
	mr.Name = chi.URLParam(r, "name")

	//type validation
	switch mr.Type {
	case "gauge":
	case "counter":
	default:
		errText := fmt.Sprintf("unexpected metric type: %s", mr.Type)
		logger.Error("handleRequestMetricV1: " + errText)
		http.Error(w, errText, http.StatusBadRequest)
		return
	}

	// val, ok := stor.Metrics[mr.Name]
	// if !ok {
	// 	http.Error(w, "Metric not found: "+mr.Name, http.StatusNotFound)
	// 	return
	// }
	metric, err := stor.GetMetric(mr.Name)
	if err != nil {
		errText := fmt.Sprintf("ERROR getting metric [%s]: %s", mr.Name, err.Error())
		logger.Error("handleRequestMetricV1: " + errText)
		http.Error(w, errText, http.StatusNotFound)
		return
	}

	//write metric value in plaintext
	body := fmt.Sprintf("%v", metric.GetValue())
	_, _ = w.Write([]byte(body))

	//w.WriteHeader(http.StatusOK)
}

// HTTP single metric request v2 processing
func handleRequestMetricV2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m domain.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		errText := fmt.Sprintf("DECODE ERROR: %s", err.Error())
		logger.Error("handleRequestMetricV2: " + errText)
		http.Error(w, errText, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(m.ID) == "" {
		errText := "DECODE ERROR: empty metric id"
		logger.Error("handleRequestMetricV2: " + errText)
		http.Error(w, "DECODE ERROR: empty metric id", http.StatusBadRequest)
		return
	}

	mr := new(domain.MetricRequest)
	mr.Type = m.MType
	mr.Name = m.ID

	//type validation
	switch mr.Type {
	case "gauge":
	case "counter":
	default:
		errText := fmt.Sprintf("unexpected metric type: %s", mr.Type)
		logger.Error("handleRequestMetricV2: " + errText)
		http.Error(w, errText, http.StatusNotFound)
		return
	}

	// metric, ok := stor.Metrics[mr.Name]
	// if !ok {
	// 	http.Error(w, "Metric not found: "+mr.Name, http.StatusNotFound)
	// 	return
	// }

	metric, err := stor.GetMetric(mr.Name)
	if err != nil {
		errText := fmt.Sprintf("ERROR getting metric [%s]: %s", mr.Name, err.Error())
		logger.Error("handleRequestMetricV2: " + errText)
		http.Error(w, errText, http.StatusNotFound)
		return
	}

	//return current metric value
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
		errText := fmt.Sprintf("ENCODE ERROR: %s", err.Error())
		logger.Error("handleRequestMetricV2: " + errText)
		http.Error(w, errText, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

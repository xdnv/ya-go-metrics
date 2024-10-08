package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"internal/app"
	"internal/domain"

	"github.com/go-chi/chi/v5"
)

// HTTP single metric update v1 processing
func handleUpdateMetricV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mr := new(domain.MetricRequest)
	mr.Type = chi.URLParam(r, "type")
	mr.Name = chi.URLParam(r, "name")
	mr.Value = chi.URLParam(r, "value")

	err := stor.UpdateMetricS(mr.Type, mr.Name, mr.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//save dump if set to immediate mode
	if (sc.StorageMode == app.File) && (sc.StoreInterval == 0) {
		err := stor.SaveState(sc.FileStoragePath)
		if err != nil {
			fmt.Printf("srv-updateMetricV1: failed to save server state to [%s], error: %s\n", sc.FileStoragePath, err)
		}
	}

	//w.WriteHeader(http.StatusOK)
}

// HTTP single metric update v2 processing
func handleUpdateMetricV2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m domain.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		fmt.Printf("DECODE ERROR: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(m.ID) == "" {
		http.Error(w, "DECODE ERROR: empty metric id", http.StatusBadRequest)
		return
	}

	mr := new(domain.MetricRequest)
	mr.Type = m.MType
	mr.Name = m.ID

	switch mr.Type {
	case "gauge":
		mr.Value = fmt.Sprint(*m.Value)
	case "counter":
		mr.Value = fmt.Sprint(*m.Delta)
	default:
		http.Error(w, fmt.Sprintf("ERROR: unsupported metric type %s", mr.Type), http.StatusNotFound)
	}

	err := stor.UpdateMetricS(mr.Type, mr.Name, mr.Value)
	if err != nil {
		fmt.Printf("UPDATE ERROR: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//save dump if set to immediate mode
	if (sc.StorageMode == app.File) && (sc.StoreInterval == 0) {
		err := stor.SaveState(sc.FileStoragePath)
		if err != nil {
			fmt.Printf("srv-updateMetricV2: failed to save server state to [%s], error: %s\n", sc.FileStoragePath, err)
		}
	}

	// metric := stor.Metrics[mr.Name]
	metric, err := stor.GetMetric(mr.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

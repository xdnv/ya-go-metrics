package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"internal/adapters/logger"
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
		errText := fmt.Sprintf("failed to update metric [%s][%s][%s], error: %s", mr.Type, mr.Name, mr.Value, err.Error())
		logger.Error("handleUpdateMetricV1: " + errText)
		http.Error(w, errText, http.StatusBadRequest)
		return
	}

	//save dump if set to immediate mode
	if (sc.StorageMode == app.File) && (sc.StoreInterval == 0) {
		err := stor.SaveState(sc.FileStoragePath)
		if err != nil {
			errText := fmt.Sprintf("failed to save server state to [%s], error: %s", sc.FileStoragePath, err.Error())
			logger.Error("handleUpdateMetricV1: " + errText)
		}
	}

	//w.WriteHeader(http.StatusOK)
}

// HTTP single metric update v2 processing
func handleUpdateMetricV2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m domain.Metrics

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		errText := fmt.Sprintf("DECODE ERROR: %s", err.Error())
		logger.Error("handleUpdateMetricV2: " + errText)
		http.Error(w, errText, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(m.ID) == "" {
		errText := "DECODE ERROR: empty metric id"
		logger.Error("handleUpdateMetricV2: " + errText)
		http.Error(w, errText, http.StatusBadRequest)
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
		errText := fmt.Sprintf("ERROR: unsupported metric type %s", mr.Type)
		logger.Error("handleUpdateMetricV2: " + errText)
		http.Error(w, errText, http.StatusNotFound)
	}

	err := stor.UpdateMetricS(mr.Type, mr.Name, mr.Value)
	if err != nil {
		errText := fmt.Sprintf("failed to update metric [%s][%s][%s], error: %s", mr.Type, mr.Name, mr.Value, err.Error())
		logger.Error("handleUpdateMetricV2: " + errText)
		http.Error(w, errText, http.StatusBadRequest)
		return
	}

	//save dump if set to immediate mode
	if (sc.StorageMode == app.File) && (sc.StoreInterval == 0) {
		err = stor.SaveState(sc.FileStoragePath)
		if err != nil {
			errText := fmt.Sprintf("failed to save server state to [%s], error: %s", sc.FileStoragePath, err.Error())
			logger.Error("handleUpdateMetricV2: " + errText)
		}
	}

	// metric := stor.Metrics[mr.Name]
	metric, err := stor.GetMetric(mr.Name)
	if err != nil {
		errText := fmt.Sprintf("failed to get metric [%s], error: %s", mr.Name, err.Error())
		logger.Error("handleUpdateMetricV2: " + errText)
		http.Error(w, errText, http.StatusInternalServerError)
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
		errText := fmt.Sprintf("ENCODE ERROR: %s", err.Error())
		logger.Error("handleUpdateMetricV2: " + errText)
		http.Error(w, errText, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

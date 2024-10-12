package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"internal/adapters/logger"
	"internal/app"
	"internal/domain"
)

// HTTP mass metric update processing
func handleUpdateMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m []domain.Metrics
	var errs []error

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		errText := fmt.Sprintf("DECODE ERROR: %s", err.Error())
		logger.Error("handleUpdateMetrics: " + errText)
		http.Error(w, errText, http.StatusBadRequest)
		return
	}

	mr := stor.BatchUpdateMetrics(&m, &errs)

	//handling all errors encountered
	if len(errs) > 0 {
		strErrors := make([]string, len(errs))
		for i, err := range errs {
			strErrors[i] = err.Error()
		}
		errText := fmt.Sprintf("bulk update errors: %s", strings.Join(strErrors, "\n"))
		logger.Error("handleUpdateMetrics: " + errText)
		http.Error(w, errText, http.StatusBadRequest)
		return
	}

	//save dump if set to immediate mode
	if (sc.StorageMode == app.File) && (sc.StoreInterval == 0) {
		err := stor.SaveState(sc.FileStoragePath)
		if err != nil {
			errText := fmt.Sprintf("failed to save server state to [%s], error: %s", sc.FileStoragePath, err.Error())
			logger.Error("handleUpdateMetrics: " + errText)
		}
	}

	// metric, err := stor.GetMetric(mr.Name)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }

	// switch mr.Type {
	// case "gauge":
	// 	metricValue := metric.GetValue().(float64)
	// 	m.Value = &metricValue
	// case "counter":
	// 	metricValue := metric.GetValue().(int64)
	// 	m.Delta = &metricValue
	// }

	resp, err := json.Marshal(mr)
	if err != nil {
		errText := fmt.Sprintf("ENCODE ERROR: %s", err.Error())
		logger.Error("handleUpdateMetrics: " + errText)
		http.Error(w, errText, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

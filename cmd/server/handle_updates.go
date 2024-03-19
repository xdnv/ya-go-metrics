package main

import (
	"encoding/json"
	"fmt"
	"internal/app"
	"internal/domain"
	"net/http"
	"strings"
)

// HTTP update processing
func updateMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var m []domain.Metrics
	var errs []error

	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		fmt.Printf("DECODE ERROR: %s", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, v := range m {
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
			errs = append(errs, err)
			//http.Error(w, fmt.Sprintf("ERROR: unsupported metric type %s", mr.Type), http.StatusNotFound)
			continue
		}

		err := stor.UpdateMetricS(mr.Type, mr.Name, mr.Value)
		if err != nil {
			fmt.Printf("UPDATE ERROR: %s", err.Error())
			errs = append(errs, err)
			continue
		}
	}

	//handling all errors encountered
	if len(errs) > 0 {
		strErrors := make([]string, len(errs))
		for i, err := range errs {
			strErrors[i] = err.Error()
		}

		http.Error(w, "Errors: \n"+strings.Join(strErrors, "\n"), http.StatusBadRequest)
		return
	}

	//save dump if set to immediate mode
	if (sc.StorageMode == app.File) && (sc.StoreInterval == 0) {
		err := stor.SaveState(sc.FileStoragePath)
		if err != nil {
			fmt.Printf("srv-updateMetrics: failed to save server state to [%s], error: %s\n", sc.FileStoragePath, err)
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

	resp, err := json.Marshal(m)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// HTTP request processing
func requestMetric(w http.ResponseWriter, r *http.Request) {
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

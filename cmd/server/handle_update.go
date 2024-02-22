package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// HTTP update processing
func updateMetric(w http.ResponseWriter, r *http.Request) {
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

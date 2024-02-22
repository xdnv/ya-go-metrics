package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// HTTP update processing
func updateMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mr := new(MetricRequest)
	// mr.Mode = "update"
	mr.Type = chi.URLParam(r, "type")
	mr.Name = chi.URLParam(r, "name")
	mr.Value = chi.URLParam(r, "value")

	err := storage.UpdateMetricS(mr.Type, mr.Name, mr.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// switch mr.Type {
	// case "gauge":
	// 	val, ok := storage.Metrics[mr.Name].(*Gauge)
	// 	if !ok {
	// 		//create new item if not exists
	// 		val = &Gauge{}
	// 	}
	// 	err := val.UpdateValueS(mr.Value)
	// 	if err != nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		return
	// 	}
	// 	storage.Metrics[mr.Name] = val
	// case "counter":
	// 	val, ok := storage.Metrics[mr.Name].(*Counter)
	// 	if !ok {
	// 		//create new item if not exists
	// 		val = &Counter{}
	// 	}
	// 	err := val.UpdateValueS(mr.Value)
	// 	if err != nil {
	// 		w.WriteHeader(http.StatusBadRequest)
	// 		return
	// 	}
	// 	storage.Metrics[mr.Name] = val
	// default:
	// 	http.Error(w, fmt.Sprintf("unexpected metric type: %s", mr.Mode), http.StatusBadRequest)
	// 	return
	// }

	//w.WriteHeader(http.StatusOK)
}

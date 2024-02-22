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
	mr.Mode = "value"
	mr.Type = chi.URLParam(r, "type")
	mr.Name = chi.URLParam(r, "name")

	//type validation
	switch mr.Type {
	case "gauge":
	case "counter":
	default:
		http.Error(w, fmt.Sprintf("unexpected metric type: %s", mr.Mode), http.StatusBadRequest)
		return
	}

	val, ok := storage.Metrics[mr.Name]
	if !ok {
		//w.WriteHeader(http.StatusNotFound)
		http.Error(w, "Metric not found: "+mr.Name, http.StatusNotFound)
		return
	}

	//write metric value in plaintext
	body := fmt.Sprintf("%v", val.GetValue())
	_, _ = w.Write([]byte(body))

	// //===========================

	// body := fmt.Sprintf("Method: %s\r\n", r.Method)
	// body += fmt.Sprintf("STORAGE: %s\r\n", storage)
	// body += fmt.Sprintf("URL: %s\r\n", r.URL)
	// body += "Header ===============\r\n"
	// for k, v := range r.Header {
	// 	body += fmt.Sprintf("%s: %v\r\n", k, v)
	// }
	// body += "Query parameters ===============\r\n"
	// for k, v := range r.URL.Query() {
	// 	body += fmt.Sprintf("%s: %v\r\n", k, v)
	// }
	// _, _ = w.Write([]byte(body))

	//w.WriteHeader(http.StatusOK)
}

package main

import (
	"net/http"

	"internal/adapters/logger"
	"internal/app"
	"internal/domain"

	"github.com/go-chi/chi/v5"
)

// HTTP single metric request v1 processing
func handleRequestMetricV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mr := new(domain.MetricRequest)
	mr.Type = chi.URLParam(r, "type")
	mr.Name = chi.URLParam(r, "name")

	hs := app.RequestMetricV1(mr)
	if hs.Err != nil {
		logger.Error("handleRequestMetricV1: " + hs.Message)
		http.Error(w, hs.Message, hs.HTTPStatus)
		return
	}

	w.Write([]byte(hs.Message))
}

// HTTP single metric request v2 processing
func handleRequestMetricV2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, hs := app.RequestMetricV2(r.Body)
	if hs.Err != nil {
		logger.Error("handleRequestMetricV2: " + hs.Message)
		http.Error(w, hs.Message, hs.HTTPStatus)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(*data)
}

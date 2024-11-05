package http_server

import (
	"net/http"

	"internal/adapters/logger"
	"internal/app"
	"internal/domain"

	"github.com/go-chi/chi/v5"
)

// HTTP single metric update v1 processing
func HandleUpdateMetricV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mr := new(domain.MetricRequest)
	mr.Type = chi.URLParam(r, "type")
	mr.Name = chi.URLParam(r, "name")
	mr.Value = chi.URLParam(r, "value")

	hs := app.UpdateMetricV1(mr)
	if hs.Err != nil {
		logger.Error("handleUpdateMetricV1: " + hs.Message)
		http.Error(w, hs.Message, hs.HTTPStatus)
		return
	}

	//w.Write([]byte(hs.Message))
}

// HTTP single metric update v2 processing
func HandleUpdateMetricV2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, hs := app.UpdateMetricV2(r.Body)
	if hs.Err != nil {
		logger.Error("handleUpdateMetricV2: " + hs.Message)
		http.Error(w, hs.Message, hs.HTTPStatus)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(*data)
}

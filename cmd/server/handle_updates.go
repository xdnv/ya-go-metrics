package main

import (
	"net/http"

	"internal/adapters/logger"
	"internal/app"
)

// HTTP mass metric update processing
func handleUpdateMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, hs := app.UpdateMetrics(r.Body)
	if hs.Err != nil {
		logger.Error("handleUpdateMetrics: " + hs.Message)
		http.Error(w, hs.Message, hs.HTTPStatus)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(*data)
}

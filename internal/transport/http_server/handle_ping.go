package http_server

import (
	"net/http"

	"internal/adapters/logger"
	"internal/app"
)

// HTTP request processing
func HandlePingDBServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//w.WriteHeader(http.StatusOK)

	hs := app.PingDBServer()
	if hs.Err != nil {
		logger.Error(hs.Message)
		http.Error(w, hs.Message, hs.HTTPStatus)
		return
	}

	w.Write([]byte(hs.Message))
}

package http_server

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"internal/adapters/cryptor"
	"internal/adapters/firewall"
	"internal/adapters/logger"
	"internal/adapters/signer"

	"internal/app"
)

// http server part
func ServeHTTP() *http.Server {
	mux := chi.NewRouter()
	mux.Use(logger.LoggerMiddleware)
	mux.Use(firewall.HandleTrustedNetworkRequests)
	mux.Use(signer.HandleSignedRequests)
	mux.Use(HandleGZIPRequests)
	mux.Use(cryptor.HandleEncryptedRequests)
	if app.Sc.CompressReplies {
		mux.Use(middleware.Compress(5, app.Sc.CompressibleContentTypes...))
	}

	mux.Get("/", HandleIndex)
	mux.Get("/ping", HandlePingDBServer)
	mux.Post("/value/", HandleRequestMetricV2)
	mux.Get("/value/{type}/{name}", HandleRequestMetricV1)
	mux.Post("/update/", HandleUpdateMetricV2)
	mux.Post("/update/{type}/{name}/{value}", HandleUpdateMetricV1)
	mux.Post("/updates/", HandleUpdateMetrics)

	mux.Mount("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	// create a server
	srv := &http.Server{Addr: app.Sc.Endpoint, Handler: mux}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			logger.Errorf("Listen: %s", err.Error())
		}
	}()

	return srv
}

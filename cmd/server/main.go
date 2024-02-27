package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type MetricRequest struct {
	Type  string
	Name  string
	Value string
}

var storage = NewMemStorage()

func main() {
	if err := run(); err != nil {
		//logger.Error("Server error", zap.Error(err))
		log.Fatal(err)
	}
}

func run() error {
	sc := InitServerConfig()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap logger")
	}
	defer logger.Sync()

	sugar := logger.Sugar()

	//sugar.Infof("Failed to fetch URL: %s", url)
	//sugar.Errorf("Failed to fetch URL: %s", url)

	//fmt.Printf("using endpoint: %s\n", sc.Endpoint)
	sugar.Infof("using endpoint: %s", sc.Endpoint)

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	mux.Get("/", index)
	mux.Get("/value/{type}/{name}", requestMetric)
	mux.Post("/update/{type}/{name}/{value}", updateMetric)

	return http.ListenAndServe(sc.Endpoint, mux)
}

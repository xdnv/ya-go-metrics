package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type MetricRequest struct {
	Type  string
	Name  string
	Value string
}

var storage = NewMemStorage()

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	sc := InitServerConfig()

	fmt.Printf("using endpoint: %s\n", sc.Endpoint)

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	mux.Get("/", index)
	mux.Get("/value/{type}/{name}", requestMetric)
	mux.Post("/update/{type}/{name}/{value}", updateMetric)

	return http.ListenAndServe(sc.Endpoint, mux)
}

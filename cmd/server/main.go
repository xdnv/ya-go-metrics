package main

//TODO: Сейчас адлгоритм позволяет указывать одно имя метрики для разных типов, в этом случае может происходить замена типа метрики и непредсказуемое поведение значения
//можно либо разделить хранение метрик по мапам для каждого типа, либо добавить признак типа метрики в саму метрику и сверять с ней, либо что-то ещё
// основная часть программы

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type MetricRequest struct {
	Mode  string
	Type  string
	Name  string
	Value string
}

var storage = NewMemStorage()

func main() {
	if err := run(); err != nil {
		// panic(err)
		log.Fatal(err)
	}
}

// init dependencies
func run() error {
	sc := InitServerConfig()

	fmt.Printf("using endpoint: %s\n", sc.Endpoint)

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	mux.Get("/", index)
	mux.Get("/value/{type}/{name}", requestMetric)
	mux.Post("/update/{type}/{name}/{value}", updateMetric)

	//log.Fatal(http.ListenAndServe(sc.Endpoint, mux))
	return http.ListenAndServe(sc.Endpoint, mux)
}

package main

//TODO: Сейчас адлгоритм позволяет указывать одно имя метрики для разных типов, в этом случае может происходить замена типа метрики и непредсказуемое поведение значения
//можно либо разделить хранение метрик по мапам для каждого типа, либо добавить признак типа метрики в саму метрику и сверять с ней, либо что-то ещё
// основная часть программы

import (
	"fmt"
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

	// это пойдёт в тесты
	g := Gauge{Value: 0.0} //new(Gauge)
	g.UpdateValue(0.011)
	g.UpdateValue(0.012)

	c := Counter{Value: 0} //new(Counter)
	c.UpdateValue(50)
	c.UpdateValue(60)

	storage.Metrics["Type1G"] = g // append
	storage.Metrics["Type2C"] = c // append

	fmt.Printf("Gauge Metric: %v\n", GetMetricValue(storage.Metrics["Type1G"]))
	fmt.Printf("Counter Metric: %v\n", GetMetricValue(storage.Metrics["Type2C"]))

	if err := run(); err != nil {
		panic(err)
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

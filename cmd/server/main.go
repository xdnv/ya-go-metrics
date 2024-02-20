package main

//TODO: Сейчас адлгоритм позволяет указывать одно имя метрики для разных типов, в этом случае может происходить замена типа метрики и непредсказуемое поведение значения
//можно либо разделить хранение метрик по мапам для каждого типа, либо добавить признак типа метрики в саму метрику и сверять с ней, либо что-то ещё
// основная часть программы
//todo:
// +++ оптимайз функций - вынести повторы в отд функции
// +++ добавить вывод значения через get
// +++ записывать все метрики
// - перевести на HTTP фреймворк
// - добавить тесты
// - в агенте разделить время получения данных и время отправки на сервер
// - модульность? абстракции? мутексы?

import (
	"cmp"
	"errors"
	"fmt"
	"io"

	"net/http"
	"slices"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type MetricRequest struct {
	Mode  string
	Type  string
	Name  string
	Value string
}

var storage = InitStorage()

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

	//standard router library
	// mux := http.NewServeMux()
	// mux.HandleFunc(`/`, index)
	// mux.HandleFunc(`/update/`, updateMetric)
	// mux.HandleFunc(`/value/`, requestMetric)

	mux := chi.NewRouter()
	mux.Use(middleware.Logger)

	mux.Get("/", index)
	mux.Get("/value/{type}/{name}", requestMetric)
	mux.Post("/update/{type}/{name}/{value}", updateMetric)

	//err := http.ListenAndServe(`:8080`, mux)
	//log.Fatal(http.ListenAndServe(sc.Endpoint, mux))
	return http.ListenAndServe(sc.Endpoint, mux)
}

func sortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// const metricsHTML = `
//  	<h1>{{.PageTitle}}</h1>
// 	 <style>
// 	 table, td, th {
// 	   border: 1px solid black;
// 	   border-spacing: 0px;
// 	 }
// 	 </style>
//  	<table>
// 		{{range .Metrics}}
// 			{{if .Header}}
// 				<tr><th>Metric</th><th>Value</th></tr>
// 			{{else}}
// 				<tr><td>{{.Title}}</td><td style=\"text-align: right;\">{{.Value}}</td></tr>
// 			{{end}}
// 		{{end}}
// 	</table>
// 	`

// type MetricEntry struct {
// 	Title  string
// 	Value  string
// 	Header bool
// }

// type MetricPageData struct {
// 	PageTitle string
// 	Metrics   []MetricEntry
// }

// func index_t(w http.ResponseWriter, r *http.Request) {

// 	//check for malformed requests
// 	if r.URL.Path != "/" {
// 		http.NotFound(w, r)
// 		return
// 	}

// 	// set correct datatype in header
// 	w.Header().Set("Content-Type", "text/html; charset=utf-8")
// 	w.WriteHeader(http.StatusOK)

// 	data := new(MetricPageData)

// 	data.PageTitle = "Current values"
// 	data.Metrics = append(data.Metrics, MetricEntry{"Metric", "Value", true})

// 	for _, key := range sortedKeys(storage.Metrics) {
// 		data.Metrics = append(data.Metrics, MetricEntry{key, fmt.Sprintf("%v", storage.Metrics[key].(Metric).GetValue()), false})
// 	}

// 	tmpl := template.Must(template.New("").Parse(metricsHTML))
// 	tmpl.Execute(w, data)
// }

func index(w http.ResponseWriter, r *http.Request) {

	//check for malformed requests - only exact root path accepted
	//covered by tests, removal will bring tests to fail
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// установим правильный заголовок для типа данных
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	const pageTmpl = `<html>
		<head>
			<title>%s</title>
			<style>
	  		table, td, th {
	    		border: 1px solid black;
	    		border-spacing: 0px;
	  		}
	  		td, th {
	    		padding: 5px;
	  		}
			</style>
		</head>
		<body>
	    	%s
		</body>
		</html>`

	tableOut := "<table>"

	headerTmpl := "<tr><th>%s</th><th>%v</th></tr>"
	rowTmpl := "<tr><td>%s</td><td style=\"text-align: right;\">%v</td></tr>"

	tableOut += fmt.Sprintf(headerTmpl, "Metric", "Value")

	for _, key := range sortedKeys(storage.Metrics) {
		tableOut += fmt.Sprintf(rowTmpl, key, storage.Metrics[key].GetValue())
	}
	tableOut += "</table>"

	io.WriteString(w, fmt.Sprintf(pageTmpl, "Metrics", tableOut))
}

// type appError struct {
// 	Error   error
// 	Message string
// 	Code    int
// }

// func extractMetricRequest(mURL string) (*MetricRequest, error) {

// 	mr := new(MetricRequest)

// 	//split path structure, extract metric value (if any)
// 	metricURL, metricValue := path.Split(mURL)

// 	//split configuration URL
// 	splitFunc := func(c rune) bool {
// 		return c == '/'
// 	}
// 	metricParams := strings.FieldsFunc(metricURL, splitFunc)

// 	// POST http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ> + metricValue == "<ЗНАЧЕНИЕ_МЕТРИКИ>"
// 	// GET /value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> + metricValue == ""

// 	//check for expected query structure
// 	numParams := len(metricParams)
// 	if numParams != 3 {
// 		//return mr, fmt.Errorf("the URL parameter quantity is %d while expected 3", numParams)
// 		//fmt.Printf("TRACE: the URL parameter quantity is %d while expected 3\n", numParams)
// 		return mr, fmt.Errorf("the URL parameter quantity is %d while expected 3", numParams)
// 	}

// 	mr.Mode = metricParams[0]
// 	mr.Type = metricParams[1]
// 	mr.Name = metricParams[2]
// 	mr.Value = metricValue

// 	return mr, nil
// }

func extractMetricRequestChi(r *http.Request, mode string) (*MetricRequest, error) {

	mr := new(MetricRequest)

	mr.Mode = mode
	mr.Type = chi.URLParam(r, "type")
	mr.Name = chi.URLParam(r, "name")
	mr.Value = ""

	if mr.Mode == "update" {
		mr.Value = chi.URLParam(r, "value")
	}

	return mr, nil
}

func validateMetricRequest(mr MetricRequest) (*MetricRequest, error) {

	//fmt.Printf("TRACE: validate mode[%s] type[%s] name[%s] value[%s]\n", mr.Mode, mr.Type, mr.Name, mr.Value)

	if !slices.Contains([]string{"update", "value"}, mr.Mode) {
		//return mr, fmt.Errorf("unexpected metric processing mode: %s", mr.Mode)
		return &mr, fmt.Errorf("unexpected metric processing mode: %s", mr.Mode)
	}

	if !slices.Contains([]string{"gauge", "counter"}, mr.Type) {
		//return mr, fmt.Errorf("unexpected metric type: %s", mr.Mode)

		return &mr, fmt.Errorf("unexpected metric type: %s", mr.Mode)
	}

	if (mr.Mode == "update") && (mr.Value == "") {
		//return mr, errors.New("non-empty metric Value parameter expected for update request")
		return &mr, errors.New("non-empty metric Value parameter expected for update request")
	}

	return &mr, nil
}

// HTTP request processing
func requestMetric(w http.ResponseWriter, r *http.Request) {

	// установим правильный заголовок для типа данных
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mr, err := extractMetricRequestChi(r, "value")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	mr, err = validateMetricRequest(*mr)
	if err != nil {
		//w.WriteHeader(http.StatusBadRequest)
		//fmt.Printf("TRACE: failed validation exit message [%s], status code [%d]\n", aerr.Error.Error(), aerr.Code)
		//http.Error(w, "Malformed request", http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	val, ok := storage.Metrics[mr.Name]
	if !ok {
		//w.WriteHeader(http.StatusNotFound)
		http.Error(w, "Metric not found: "+mr.Name, http.StatusNotFound)
		return
	}

	body := fmt.Sprintf("%v", val.GetValue())
	_, _ = w.Write([]byte(body))

	// это пойдёт в тесты
	// //===========================

	// body := fmt.Sprintf("Method: %s\r\n", r.Method)

	// body += fmt.Sprintf("STORAGE: %s\r\n", storage)

	// body += fmt.Sprintf("URL: %s\r\n", r.URL)
	// // /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	// // /update/ = мы уже здесь благодаря обработчику, удаляем
	// // <ТИП_МЕТРИКИ> = Gauge | Counter
	// // <ИМЯ_МЕТРИКИ> = произвольное, пока не изучен пакет "runtime"
	// // <ЗНАЧЕНИЕ_МЕТРИКИ> = float для Gauge, int для Counter

	// body += "Header ===============\r\n"
	// for k, v := range r.Header {
	// 	body += fmt.Sprintf("%s: %v\r\n", k, v)
	// }
	// body += "Query parameters ===============\r\n"
	// for k, v := range r.URL.Query() {
	// 	body += fmt.Sprintf("%s: %v\r\n", k, v)
	// }
	// _, _ = w.Write([]byte(body))

	// // пока установим ответ-заглушку, без проверки ошибок
	// // _, _ = w.Write([]byte(`
	// //   {
	// //     "response": {
	// //       "text": "Извините, я пока ничего не умею"
	// //     },
	// //     "version": "1.0"
	// //   }
	// // `))

	//w.WriteHeader(http.StatusOK)
}

// HTTP request processing
func updateMetric(w http.ResponseWriter, r *http.Request) {

	// установим правильный заголовок для типа данных
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mr, err := extractMetricRequestChi(r, "update")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	mr, err = validateMetricRequest(*mr)
	if err != nil {
		//w.WriteHeader(http.StatusBadRequest)
		//fmt.Printf("TRACE: failed validation exit message [%s], status code [%d]\n", aerr.Error.Error(), aerr.Code)
		//http.Error(w, "Malformed request", http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch mr.Type {
	case "gauge":
		val, ok := storage.Metrics[mr.Name].(Gauge)
		if !ok {
			//создаём новый элемент
			val = Gauge{}
		}
		err := val.UpdateValueS(mr.Value)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storage.Metrics[mr.Name] = val
	case "counter":
		val, ok := storage.Metrics[mr.Name].(Counter)
		if !ok {
			//создаём новый элемент
			val = Counter{}
		}
		err := val.UpdateValueS(mr.Value)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storage.Metrics[mr.Name] = val
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//w.WriteHeader(http.StatusOK)
}

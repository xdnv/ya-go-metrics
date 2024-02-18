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
	"path"
	"slices"
	"sort"
	"strings"
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

	fmt.Printf("Gauge Metric: %v\n", GetMetricValue(storage.Metrics["Type1G"].(Metric)))
	fmt.Printf("Counter Metric: %v\n", GetMetricValue(storage.Metrics["Type2C"].(Metric)))

	if err := run(); err != nil {
		panic(err)
	}
}

// init dependencies
func run() error {
	sc := InitServerConfig()

	fmt.Printf("using endpoint: %s\n", sc.Endpoint)

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, index)
	mux.HandleFunc(`/update/`, updateMetric)
	mux.HandleFunc(`/value/`, requestMetric)

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

func index(w http.ResponseWriter, r *http.Request) {

	//check for malformed requests
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

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
		tableOut += fmt.Sprintf(rowTmpl, key, storage.Metrics[key].(Metric).GetValue())
	}
	tableOut += "</table>"

	io.WriteString(w, fmt.Sprintf(pageTmpl, "Metrics", tableOut))
}

type appError struct {
	Error   error
	Message string
	Code    int
}

func extractValidateMetricRequest(mURL string) (*MetricRequest, *appError) {

	mr := new(MetricRequest)

	//split path structure, extract metric value (if any)
	metricURL, metricValue := path.Split(mURL)

	//split configuration URL
	splitFunc := func(c rune) bool {
		return c == '/'
	}
	metricParams := strings.FieldsFunc(metricURL, splitFunc)

	// POST http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ> + metricValue == "<ЗНАЧЕНИЕ_МЕТРИКИ>"
	// GET /value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> + metricValue == ""

	//check for expected query structure
	numParams := len(metricParams)
	if numParams != 3 {
		//return mr, fmt.Errorf("the URL parameter quantity is %d while expected 3", numParams)
		fmt.Printf("TRACE: the URL parameter quantity is %d while expected 3\n", numParams)
		return mr, &appError{fmt.Errorf("the URL parameter quantity is %d while expected 3", numParams), "", http.StatusNotFound}
	}

	mr.Mode = metricParams[0]
	mr.Type = metricParams[1]
	mr.Name = metricParams[2]
	mr.Value = metricValue

	fmt.Printf("TRACE: mode[%s] type[%s] name[%s] value[%s]\n", mr.Mode, mr.Type, mr.Name, mr.Value)

	if !slices.Contains([]string{"update", "value"}, mr.Mode) {
		//return mr, fmt.Errorf("unexpected metric processing mode: %s", mr.Mode)
		return mr, &appError{fmt.Errorf("unexpected metric processing mode: %s", mr.Mode), "", http.StatusBadRequest}
	}

	if !slices.Contains([]string{"gauge", "counter"}, mr.Type) {
		//return mr, fmt.Errorf("unexpected metric type: %s", mr.Mode)
		return mr, &appError{fmt.Errorf("unexpected metric type: %s", mr.Mode), "", http.StatusBadRequest}
	}

	if (mr.Mode == "update") && (metricValue == "") {
		//return mr, errors.New("non-empty metric Value parameter expected for update request")
		return mr, &appError{errors.New("non-empty metric Value parameter expected for update request"), "", http.StatusBadRequest}
	}

	return mr, nil
}

// HTTP request processing
func requestMetric(w http.ResponseWriter, r *http.Request) {

	// разрешаем только GET-запросы
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Доработайте сервер так, чтобы в ответ на запрос GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> он возвращал текущее значение метрики в текстовом виде со статусом http.StatusOK.
	// При успешном приёме возвращать http.StatusOK.
	// При попытке запроса неизвестной метрики сервер должен возвращать http.StatusNotFound.

	// установим правильный заголовок для типа данных
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mr, aerr := extractValidateMetricRequest(r.URL.Path)
	if aerr != nil {
		//w.WriteHeader(http.StatusBadRequest)
		fmt.Printf("TRACE: failed validation exit message [%s], status code [%d]\n", aerr.Error.Error(), aerr.Code)
		//http.Error(w, "Malformed request", http.StatusBadRequest)
		http.Error(w, aerr.Error.Error(), aerr.Code)
		return
	}

	val, ok := storage.Metrics[mr.Name]
	if !ok {
		//w.WriteHeader(http.StatusNotFound)
		http.Error(w, "Metric not found: "+mr.Name, http.StatusNotFound)
		return
	}

	body := fmt.Sprintf("%v", val.(Metric).GetValue())
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

	// разрешаем только POST-запросы
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// (1.3) Принимать данные в формате http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>, Content-Type: text/plain.
	// contentType := r.Header.Get("Content-type")
	// if contentType != "text/plain" {
	// 	w.WriteHeader(http.StatusUnsupportedMediaType)
	// 	return
	// }

	// (1.4) При успешном приёме возвращать http.StatusOK.
	//При попытке передать запрос без имени метрики возвращать http.StatusNotFound.
	//При попытке передать запрос с некорректным типом метрики или значением возвращать http.StatusBadRequest.

	// установим правильный заголовок для типа данных
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	mr, err := extractValidateMetricRequest(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
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

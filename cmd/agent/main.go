package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// type GaugeValues struct {
// 	Alloc,
// 	BuckHashSys,
// 	Frees,
// 	GCCPUFraction,
// 	GCSys,
// 	HeapAlloc,
// 	HeapIdle,
// 	HeapInuse,
// 	HeapObjects,
// 	HeapReleased,
// 	HeapSys,
// 	LastGC,
// 	Lookups,
// 	MCacheInuse,
// 	MCacheSys,
// 	MSpanInuse,
// 	MSpanSys,
// 	Mallocs,
// 	NextGC,
// 	NumForcedGC,
// 	NumGC,
// 	OtherSys,
// 	PauseTotalNs,
// 	StackInuse,
// 	StackSys,
// 	Sys,
// 	TotalAlloc,
// 	RandomValue float64 //(тип gauge) — обновляемое произвольное значение.
// }

// type CounterValues struct {
// 	PollCount int64 //(тип counter) — счётчик, увеличивающийся на 1 при каждом обновлении метрики из пакета runtime (на каждый pollInterval — см. ниже).
// }

type MetricStorage struct {
	sync.RWMutex
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMetricStorage() *MetricStorage {
	var ms MetricStorage
	ms.Gauge = make(map[string]float64)
	ms.Counter = make(map[string]int64)
	return &ms
}

const floatPrecision = 1000000

func GetRandInt(min, max int) int {
	nBig, _ := rand.Int(rand.Reader, big.NewInt(int64(max+1-min)))
	n := nBig.Int64()
	return int(n) + min
}

func GetRandFloat(min, max float64) float64 {
	minInt := int(min * floatPrecision)
	maxInt := int(max * floatPrecision)

	return float64(GetRandInt(minInt, maxInt)) / floatPrecision
}

func PostValue(endpoint string, counterType string, counterName string, value string) (*http.Response, error) {
	// data := []byte(`{"foo":"bar"}`)
	// r := bytes.NewReader(data)
	address := fmt.Sprintf("http://%s/update/%s/%s/%s", endpoint, counterType, counterName, value)
	resp, err := http.Post(address, "text/plain", nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func collector(wg *sync.WaitGroup, duration int64, ac AgentConfig) {
	var rtm runtime.MemStats
	var interval = time.Duration(duration) * time.Second

	//execute to exit wait group
	defer wg.Done()

	for {
		<-time.After(interval)

		fmt.Printf("TRACE: collect metrics [%s]\n", time.Now().Format("2006-01-02 15:04:05"))

		ms.Lock()

		ms.Counter["PollCount"]++

		ms.Gauge["RandomValue"] = GetRandFloat(0.0, 30.0)

		// Read full mem stats
		runtime.ReadMemStats(&rtm)

		// Number of goroutines
		// m.NumGoroutine = runtime.NumGoroutine()

		// Misc memory stats
		ms.Gauge["Alloc"] = float64(rtm.Alloc)
		ms.Gauge["BuckHashSys"] = float64(rtm.BuckHashSys)
		ms.Gauge["Frees"] = float64(rtm.Frees)
		ms.Gauge["GCCPUFraction"] = float64(rtm.GCCPUFraction)
		ms.Gauge["GCSys"] = float64(rtm.GCSys)
		ms.Gauge["HeapAlloc"] = float64(rtm.HeapAlloc)
		ms.Gauge["HeapIdle"] = float64(rtm.HeapIdle)
		ms.Gauge["HeapInuse"] = float64(rtm.HeapInuse)
		ms.Gauge["HeapObjects"] = float64(rtm.HeapObjects)
		ms.Gauge["HeapReleased"] = float64(rtm.HeapReleased)
		ms.Gauge["HeapSys"] = float64(rtm.HeapSys)
		ms.Gauge["LastGC"] = float64(rtm.LastGC)
		ms.Gauge["Lookups"] = float64(rtm.Lookups)
		ms.Gauge["MCacheInuse"] = float64(rtm.MCacheInuse)
		ms.Gauge["MCacheSys"] = float64(rtm.MCacheSys)
		ms.Gauge["MSpanInuse"] = float64(rtm.MSpanInuse)
		ms.Gauge["MSpanSys"] = float64(rtm.MSpanSys)
		ms.Gauge["Mallocs"] = float64(rtm.Mallocs)
		ms.Gauge["NextGC"] = float64(rtm.NextGC)
		ms.Gauge["NumForcedGC"] = float64(rtm.NumForcedGC)
		ms.Gauge["NumGC"] = float64(rtm.NumGC) // GC Stats
		ms.Gauge["OtherSys"] = float64(rtm.OtherSys)
		ms.Gauge["PauseTotalNs"] = float64(rtm.PauseTotalNs) // GC Stats
		ms.Gauge["StackInuse"] = float64(rtm.StackInuse)
		ms.Gauge["StackSys"] = float64(rtm.StackSys)
		ms.Gauge["Sys"] = float64(rtm.Sys)
		ms.Gauge["TotalAlloc"] = float64(rtm.TotalAlloc)

		// Live objects = Mallocs - Frees
		// ms.LiveObjects = m.Mallocs - m.Frees

		ms.Unlock()
	}

}

func reporter(wg *sync.WaitGroup, duration int64, ac AgentConfig) {
	var interval = time.Duration(duration) * time.Second

	//execute to exit wait group
	defer wg.Done()

	for {
		<-time.After(interval)

		fmt.Printf("TRACE: send metrics [%s]\n", time.Now().Format("2006-01-02 15:04:05"))

		sendPayload(ac.Endpoint, ms)

		// metricName := "PollCount"
		// resp, err := PostValue(ac.Endpoint, "counter", metricName, fmt.Sprint(m.PollCount))
		// if err != nil {
		// 	fmt.Printf("ERROR posting value: %s, %s", metricName, err)
		// }
		// fmt.Println("response Status:", resp.Status)
		// fmt.Println("response Headers:", resp.Header)
		// body, _ := ioutil.ReadAll(resp.Body)
		// fmt.Println("response Body:", string(body))

		// // Just encode to json and print
		// b, _ := json.Marshal(m)
		// fmt.Println(string(b))
	}

}

func sendPayload(endpoint string, m *MetricStorage) {

	m.RLock()
	defer m.RUnlock()

	for k, v := range m.Gauge {
		sendMetric(endpoint, "gauge", k, fmt.Sprint(v))
	}

	for k, v := range m.Counter {
		sendMetric(endpoint, "counter", k, fmt.Sprint(v))
	}

}

func sendMetric(endpoint string, metricType string, metricName string, metricValue string) {

	resp, err := PostValue(endpoint, metricType, metricName, fmt.Sprint(metricValue))
	//_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	if err != nil {
		fmt.Printf("ERROR posting value: %s, %s", metricName, err)
	}
	if resp.StatusCode != 200 {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
	}

}

// global metric storage
var ms = NewMetricStorage()

func main() {
	ac := InitAgentConfig()
	fmt.Printf("using endpoint: %s\n", ac.Endpoint)
	fmt.Printf("poll interval: %d\n", ac.PollInterval)
	fmt.Printf("report interval: %d\n", ac.ReportInterval)

	var wg sync.WaitGroup
	wg.Add(1)
	go collector(&wg, ac.PollInterval, ac)
	go reporter(&wg, ac.ReportInterval, ac)
	wg.Wait()

}

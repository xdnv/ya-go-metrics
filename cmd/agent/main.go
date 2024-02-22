package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type MetricStorage struct {
	sync.RWMutex
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMetricStorage() *MetricStorage {
	var ms MetricStorage
	ms.Gauges = make(map[string]float64)
	ms.Counters = make(map[string]int64)
	return &ms
}

func PostValue(endpoint string, counterType string, counterName string, value string) (*http.Response, error) {
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

		// Read full mem stats
		runtime.ReadMemStats(&rtm)

		ms.Lock()

		ms.Counters["PollCount"]++

		ms.Gauges["RandomValue"] = rand.Float64()

		// Number of goroutines
		// m.NumGoroutine = runtime.NumGoroutine()

		// Misc memory stats
		ms.Gauges["Alloc"] = float64(rtm.Alloc)
		ms.Gauges["BuckHashSys"] = float64(rtm.BuckHashSys)
		ms.Gauges["Frees"] = float64(rtm.Frees)
		ms.Gauges["GCCPUFraction"] = float64(rtm.GCCPUFraction)
		ms.Gauges["GCSys"] = float64(rtm.GCSys)
		ms.Gauges["HeapAlloc"] = float64(rtm.HeapAlloc)
		ms.Gauges["HeapIdle"] = float64(rtm.HeapIdle)
		ms.Gauges["HeapInuse"] = float64(rtm.HeapInuse)
		ms.Gauges["HeapObjects"] = float64(rtm.HeapObjects)
		ms.Gauges["HeapReleased"] = float64(rtm.HeapReleased)
		ms.Gauges["HeapSys"] = float64(rtm.HeapSys)
		ms.Gauges["LastGC"] = float64(rtm.LastGC)
		ms.Gauges["Lookups"] = float64(rtm.Lookups)
		ms.Gauges["MCacheInuse"] = float64(rtm.MCacheInuse)
		ms.Gauges["MCacheSys"] = float64(rtm.MCacheSys)
		ms.Gauges["MSpanInuse"] = float64(rtm.MSpanInuse)
		ms.Gauges["MSpanSys"] = float64(rtm.MSpanSys)
		ms.Gauges["Mallocs"] = float64(rtm.Mallocs)
		ms.Gauges["NextGC"] = float64(rtm.NextGC)
		ms.Gauges["NumForcedGC"] = float64(rtm.NumForcedGC)
		ms.Gauges["NumGC"] = float64(rtm.NumGC) // GC Stats
		ms.Gauges["OtherSys"] = float64(rtm.OtherSys)
		ms.Gauges["PauseTotalNs"] = float64(rtm.PauseTotalNs) // GC Stats
		ms.Gauges["StackInuse"] = float64(rtm.StackInuse)
		ms.Gauges["StackSys"] = float64(rtm.StackSys)
		ms.Gauges["Sys"] = float64(rtm.Sys)
		ms.Gauges["TotalAlloc"] = float64(rtm.TotalAlloc)

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
	}
}

func sendPayload(endpoint string, m *MetricStorage) {
	m.RLock()
	defer m.RUnlock()

	for k, v := range m.Gauges {
		_, _ = sendMetric(endpoint, "gauge", k, fmt.Sprint(v))
	}

	for k, v := range m.Counters {
		_, err := sendMetric(endpoint, "counter", k, fmt.Sprint(v))
		//reset counter after successful transefer
		if err == nil {
			m.Counters[k] = 0
		}
	}
}

func sendMetric(endpoint string, metricType string, metricName string, metricValue string) (*http.Response, error) {
	resp, err := PostValue(endpoint, metricType, metricName, metricValue)
	resp.Body.Close()

	if err != nil {
		fmt.Printf("ERROR posting value: %s, %s", metricName, err)
	}
	if resp.StatusCode != 200 {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
	}
	return resp, err
}

// global metric storage
var ms = NewMetricStorage()

func main() {
	ac := InitAgentConfig()
	fmt.Printf("using endpoint: %s\n", ac.Endpoint)
	fmt.Printf("poll interval: %d\n", ac.PollInterval)
	fmt.Printf("report interval: %d\n", ac.ReportInterval)

	var wg sync.WaitGroup
	wg.Add(2)
	go collector(&wg, ac.PollInterval, ac)
	go reporter(&wg, ac.ReportInterval, ac)
	wg.Wait()
}

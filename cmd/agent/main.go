package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
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

func collector(ac AgentConfig, ctx context.Context, wg *sync.WaitGroup) {
	//execute to exit wait group
	defer wg.Done()

	var rtm runtime.MemStats

	ticker := time.NewTicker(time.Duration(ac.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			fmt.Printf("TRACE: collect metrics [%s]\n", now.Format("2006-01-02 15:04:05"))

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
		case <-ctx.Done():
			fmt.Println("agent-collector: stop requested")
			return
		}
	}
}

func reporter(ac AgentConfig, ctx context.Context, wg *sync.WaitGroup) {
	//execute to exit wait group
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(ac.ReportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			fmt.Printf("TRACE: send metrics [%s]\n", now.Format("2006-01-02 15:04:05"))
			sendPayload(ac.Endpoint, ms)
		case <-ctx.Done():
			fmt.Println("agent-reporter: stop requested")
			return
		}
	}
}

func sendPayload(endpoint string, m *MetricStorage) {
	m.RLock()
	defer m.RUnlock()

	for k, v := range m.Gauges {
		resp, _ := sendMetric(endpoint, "gauge", k, fmt.Sprint(v))
		resp.Body.Close()
	}

	for k, v := range m.Counters {
		resp, err := sendMetric(endpoint, "counter", k, fmt.Sprint(v))
		resp.Body.Close()
		//reset counter after successful transefer
		if err == nil {
			m.Counters[k] = 0
		}
	}
}

func sendMetric(endpoint string, metricType string, metricName string, metricValue string) (*http.Response, error) {
	resp, err := PostValue(endpoint, metricType, metricName, metricValue)

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
	// create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// a WaitGroup for the goroutines to tell us they've stopped
	wg := sync.WaitGroup{}

	wg.Add(1)
	go agent(ctx, &wg)

	// listen for ^C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	fmt.Println("agent: received ^C - shutting down")

	// tell the goroutines to stop
	fmt.Println("agent: telling goroutines to stop")
	cancel()

	// and wait for them to reply back
	wg.Wait()
	fmt.Println("agent: shutdown")
}

func agent(ctx context.Context, wg *sync.WaitGroup) {
	//execute to exit wait group
	defer wg.Done()

	ac := InitAgentConfig()
	fmt.Printf("agent: using endpoint %s\n", ac.Endpoint)
	fmt.Printf("agent: poll interval %d\n", ac.PollInterval)
	fmt.Printf("agent: report interval %d\n", ac.ReportInterval)

	wg.Add(1)
	go collector(ac, ctx, wg)
	wg.Add(1)
	go reporter(ac, ctx, wg)

	<-ctx.Done()
	fmt.Println("agent: shutdown requested")

	// shut down gracefully with timeout of 5 seconds max
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("agent: stopped")
}

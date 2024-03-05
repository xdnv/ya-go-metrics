package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"internal/app"
	"internal/domain"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"
)

var ac app.AgentConfig

func PostValueV1(ac app.AgentConfig, counterType string, counterName string, value string) (*http.Response, error) {
	address := fmt.Sprintf("http://%s/update/%s/%s/%s", ac.Endpoint, counterType, counterName, value)
	resp, err := http.Post(address, "text/plain", nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func PostValueV2(ac app.AgentConfig, body *bytes.Buffer) (*http.Response, error) {
	contentType := "application/json"

	address := fmt.Sprintf("http://%s/update/", ac.Endpoint)

	//older API
	if !ac.UseCompression {
		resp, err := http.Post(address, contentType, body)
		if err != nil {
			return resp, err
		}
		return resp, nil
	}

	var buf bytes.Buffer

	g := gzip.NewWriter(&buf)
	if _, err := g.Write(body.Bytes()); err != nil {
		return nil, err
	}
	if err := g.Close(); err != nil {
		return nil, err
	}
	r, err := http.NewRequest("POST", address, &buf)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", contentType)
	r.Header.Set("Content-Encoding", "gzip")
	r.Header.Set("Accept-Encoding", "gzip")
	resp, err := http.DefaultClient.Do(r)

	return resp, err
}

func collector(ctx context.Context, ac app.AgentConfig, wg *sync.WaitGroup) {
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

			ms.Gauges["NumGoroutine"] = float64(runtime.NumGoroutine()) // Number of goroutines
			ms.Gauges["LiveObjects"] = float64(rtm.Mallocs - rtm.Frees) // Live objects = Mallocs - Frees

			ms.Unlock()
		case <-ctx.Done():
			fmt.Println("agent-collector: stop requested")
			return
		}
	}
}

func reporter(ctx context.Context, ac app.AgentConfig, wg *sync.WaitGroup) {
	//execute to exit wait group
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(ac.ReportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			fmt.Printf("TRACE: send metrics [%s]\n", now.Format("2006-01-02 15:04:05"))
			sendPayload(ac, ms)
		case <-ctx.Done():
			fmt.Println("agent-reporter: stop requested")
			return
		}
	}
}

func sendPayload(ac app.AgentConfig, m *domain.MetricStorage) {
	m.RLock()
	defer m.RUnlock()

	for k, v := range m.Gauges {
		resp, err := sendMetric(ac, "gauge", k, &v)
		if err == nil {
			resp.Body.Close()
		}
	}

	for k, v := range m.Counters {
		resp, err := sendMetric(ac, "counter", k, &v)
		//reset counter after successful transefer
		if err == nil {
			resp.Body.Close()
			m.Counters[k] = 0
		}
	}
}

func sendMetric(ac app.AgentConfig, metricType string, metricName string, metricValue interface{}) (*http.Response, error) {
	var resp *http.Response
	var err error

	switch ac.APIVersion {
	case "v1":
		resp, err = PostValueV1(ac, metricType, metricName, fmt.Sprint(&metricValue))
	case "v2":
		var m domain.Metrics

		m.MType = metricType
		m.ID = metricName

		switch metricType {
		case "gauge":
			m.Value = metricValue.(*float64)
		case "counter":
			m.Delta = metricValue.(*int64)
		default:
			fmt.Printf("ERROR: unsupported metric type [%s]\n", metricType)
		}

		jsonres, jsonerr := json.Marshal(m)
		if jsonerr != nil {
			fmt.Printf("ERROR: JSON marshaling failed [%s]\n", jsonerr)
			return nil, jsonerr
		}

		buf := bytes.NewBuffer(jsonres)

		fmt.Printf("TRACE: POST body %s\n", buf)

		resp, err = PostValueV2(ac, buf)
	default:
		fmt.Printf("ERROR: unsupported API version %s", ac.APIVersion)
	}

	if err != nil {
		fmt.Printf("ERROR posting value: %s, %s\n", metricName, err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
	}
	return resp, err
}

// global metric storage
var ms = domain.NewMetricStorage()

func main() {
	// create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// a WaitGroup for the goroutines to tell us they've stopped
	wg := sync.WaitGroup{}

	//Warning! do not run outside function, it will break tests due to flag.Parse()
	ac = app.InitAgentConfig()

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

	fmt.Printf("agent: using endpoint %s\n", ac.Endpoint)
	fmt.Printf("agent: poll interval %d\n", ac.PollInterval)
	fmt.Printf("agent: report interval %d\n", ac.ReportInterval)

	wg.Add(1)
	go collector(ctx, ac, wg)
	wg.Add(1)
	go reporter(ctx, ac, wg)

	<-ctx.Done()
	fmt.Println("agent: shutdown requested")

	// // shut down gracefully with timeout of 5 seconds max
	// _, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	fmt.Println("agent: stopped")
}

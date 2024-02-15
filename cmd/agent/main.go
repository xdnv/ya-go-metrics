package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"reflect"
	"runtime"
	"time"
	//"github.com/go-chi/chi/v5"
	//"github.com/go-chi/chi/v5/middleware"
	//"github.com/gorilla/mux"
	//"github.com/julienschmidt/HttpRouter"
)

type Monitor struct {
	// Alloc,
	// TotalAlloc,
	// Sys,
	// Mallocs,
	// Frees,
	// LiveObjects uint64
	// PauseTotalNs uint64

	// NumGC        uint32
	// NumGoroutine int

	Alloc,
	BuckHashSys,
	Frees,
	GCCPUFraction,
	GCSys,
	HeapAlloc,
	HeapIdle,
	HeapInuse,
	HeapObjects,
	HeapReleased,
	HeapSys,
	LastGC,
	Lookups,
	MCacheInuse,
	MCacheSys,
	MSpanInuse,
	MSpanSys,
	Mallocs,
	NextGC,
	NumForcedGC,
	NumGC,
	OtherSys,
	PauseTotalNs,
	StackInuse,
	StackSys,
	Sys,
	TotalAlloc,
	RandomValue float64 //(тип gauge) — обновляемое произвольное значение.

	PollCount int64 //(тип counter) — счётчик, увеличивающийся на 1 при каждом обновлении метрики из пакета runtime (на каждый pollInterval — см. ниже).
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

// IfThenElse evaluates a condition, if true returns the first parameter otherwise the second
func IfThenElse(condition bool, a interface{}, b interface{}) interface{} {
	if condition {
		return a
	}
	return b
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

func NewMonitor(duration int64, ac AgentConfig) {
	var m Monitor
	var rtm runtime.MemStats
	var interval = time.Duration(duration) * time.Second

	for {
		<-time.After(interval)

		m.PollCount++
		m.RandomValue = GetRandFloat(0.0, 30.0)

		// Read full mem stats
		runtime.ReadMemStats(&rtm)

		// Number of goroutines
		// m.NumGoroutine = runtime.NumGoroutine()

		// Misc memory stats
		m.Alloc = float64(rtm.Alloc)
		m.BuckHashSys = float64(rtm.BuckHashSys)
		m.Frees = float64(rtm.Frees)
		m.GCCPUFraction = float64(rtm.GCCPUFraction)
		m.GCSys = float64(rtm.GCSys)
		m.HeapAlloc = float64(rtm.HeapAlloc)
		m.HeapIdle = float64(rtm.HeapIdle)
		m.HeapInuse = float64(rtm.HeapInuse)
		m.HeapObjects = float64(rtm.HeapObjects)
		m.HeapReleased = float64(rtm.HeapReleased)
		m.HeapSys = float64(rtm.HeapSys)
		m.LastGC = float64(rtm.LastGC)
		m.Lookups = float64(rtm.Lookups)
		m.MCacheInuse = float64(rtm.MCacheInuse)
		m.MCacheSys = float64(rtm.MCacheSys)
		m.MSpanInuse = float64(rtm.MSpanInuse)
		m.MSpanSys = float64(rtm.MSpanSys)
		m.Mallocs = float64(rtm.Mallocs)
		m.NextGC = float64(rtm.NextGC)
		m.NumForcedGC = float64(rtm.NumForcedGC)
		m.NumGC = float64(rtm.NumGC) // GC Stats
		m.OtherSys = float64(rtm.OtherSys)
		m.PauseTotalNs = float64(rtm.PauseTotalNs) // GC Stats
		m.StackInuse = float64(rtm.StackInuse)
		m.StackSys = float64(rtm.StackSys)
		m.Sys = float64(rtm.Sys)
		m.TotalAlloc = float64(rtm.TotalAlloc)

		// Live objects = Mallocs - Frees
		// m.LiveObjects = m.Mallocs - m.Frees

		sendPayload(ac.Endpoint, m)

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

func sendPayload(endpoint string, m Monitor) {

	s := reflect.ValueOf(&m).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		metricName := typeOfT.Field(i).Name
		//dataType := f.Type()
		metricType := ""
		metricValue := f.Interface()

		switch any(metricValue).(type) {
		case int64:
			metricType = "counter"
		case float64:
			metricType = "gauge"
		default:
			continue
		}

		//fmt.Printf("%d: %s %s = %v\n", i, metricName, dataType, metricValue)

		resp, err := PostValue(endpoint, metricType, metricName, fmt.Sprint(metricValue))
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if err != nil {
			fmt.Printf("ERROR posting value: %s, %s", metricName, err)
		}
		if resp.StatusCode != 200 {
			fmt.Println("response Status:", resp.Status)
			fmt.Println("response Headers:", resp.Header)
		}

	}
}

func main() {
	ac := InitAgentConfig()
	fmt.Printf("using endpoint: %s\n", ac.Endpoint)
	fmt.Printf("poll interval: %d\n", ac.PollInterval)
	fmt.Printf("report interval: %d\n", ac.ReportInterval)

	NewMonitor(ac.ReportInterval, ac)
}

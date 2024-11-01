// the main agent module provides agent (metric sender) function
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"internal/adapters/cryptor"
	"internal/adapters/logger"
	"internal/adapters/signer"
	"internal/app"
	"internal/domain"

	"github.com/google/uuid"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

var ac app.AgentConfig
var sendJobs chan uuid.UUID

// statically linked variables (YP iter20 requirement)
var buildVersion string
var buildDate string
var buildCommit string

// HTTP post metric value using API v1
func PostValueV1(ctx context.Context, ac app.AgentConfig, counterType string, counterName string, value string) (*http.Response, error) {
	address := fmt.Sprintf("http://%s/update/%s/%s/%s", ac.Endpoint, counterType, counterName, value)
	resp, err := http.Post(address, "text/plain", nil)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// HTTP post metric value using API v2
func PostValueV2(ctx context.Context, ac app.AgentConfig, body *bytes.Buffer) (*http.Response, error) {
	contentType := "application/json"

	address := fmt.Sprintf("http://%s/updates/", ac.Endpoint)
	if !ac.BulkUpdate {
		address = fmt.Sprintf("http://%s/update/", ac.Endpoint)
	}

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

	encrypted, err := encryptMessage(&buf)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest("POST", address, &buf)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	r.Close = true //whether to close the connection after replying to this request (for servers) or after sending the request (for clients).

	r.Header.Set("Content-Type", contentType)
	r.Header.Set("Content-Encoding", "gzip")
	r.Header.Set("Accept-Encoding", "gzip")

	if encrypted {
		r.Header.Set("X-Encrypted", "true")
	}

	signMessage(r, body) //body has to be signed before compression since server checks signature of unpacked data
	resp, err := http.DefaultClient.Do(r)

	return resp, err
}

func signMessage(r *http.Request, body *bytes.Buffer) error {
	if !signer.UseSignedMessaging() {
		return nil
	}

	sig, err := signer.GetSignature(body.Bytes())
	if err != nil {
		return err
	}

	//hmac.Equal
	r.Header.Set(signer.GetSignatureToken(), base64.URLEncoding.EncodeToString(sig))

	return nil
}

func encryptMessage(body *bytes.Buffer) (bool, error) {
	if !cryptor.CanEncrypt() {
		return false, nil
	}

	msg, err := cryptor.Encrypt(body.Bytes())
	if err != nil {
		return false, err
	}
	body.Reset()    // Clear the buffer
	body.Write(msg) // Write encrypted data back to buffer

	return true, nil
}

func collector(ctx context.Context, ac app.AgentConfig, wg *sync.WaitGroup) {
	//execute to exit wait group
	defer wg.Done()

	var rtm runtime.MemStats

	ticker := time.NewTicker(time.Duration(ac.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("agent-collector: collect metrics")

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

			//moved to separate goroutine
			// // add metrics from gopsutil
			// vm, err := mem.VirtualMemory()
			// if err == nil {
			// 	ms.Gauges["TotalMemory"] = float64(vm.Total)
			// 	ms.Gauges["FreeMemory"] = float64(vm.Free)
			// }

			// percentage, err := cpu.Percent(0, true)
			// if err == nil {
			// 	for idx, cpupercent := range percentage {
			// 		ms.Gauges[fmt.Sprintf("CPUutilization%d", idx)] = float64(cpupercent)
			// 	}
			// }

			ms.Unlock()
		case <-ctx.Done():
			logger.Info("agent-collector: stop requested")
			return
		}
	}
}

func collectorPs(ctx context.Context, ac app.AgentConfig, wg *sync.WaitGroup) {
	//execute to exit wait group
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(ac.PollInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("agent-collectorPs: collect PS metrics")

			// add metrics from gopsutil
			vm, err := mem.VirtualMemory()

			ms.Lock()

			if err == nil {
				ms.Gauges["TotalMemory"] = float64(vm.Total)
				ms.Gauges["FreeMemory"] = float64(vm.Free)
			}

			percentage, err := cpu.Percent(0, true)
			if err == nil {
				for idx, cpupercent := range percentage {
					ms.Gauges[fmt.Sprintf("CPUutilization%d", idx)] = float64(cpupercent)
				}
			}

			ms.Unlock()
		case <-ctx.Done():
			logger.Info("agent-collector-ps: stop requested")
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
		case <-ticker.C:
			if !ac.UseRateLimit {
				logger.Info("agent-reporter: send metrics")
				sendPayload(ctx, ac, ms)
			} else {
				logger.Info("agent-reporter: send metrics rated")
				sendJobs <- uuid.New()
			}
		case <-ctx.Done():
			logger.Info("agent-reporter: stop requested")
			return
		}
	}
}

func sendPayload(ctx context.Context, ac app.AgentConfig, m *domain.MetricStorage) {
	m.RLock()
	defer m.RUnlock()

	var ma []domain.Metrics

	for k, v := range m.Gauges {
		m := domain.Metrics{MType: "gauge", ID: k}
		val := v //needs to be local
		m.Value = &val
		ma = append(ma, m)
	}

	for k, v := range m.Counters {
		m := domain.Metrics{MType: "counter", ID: k}
		val := v //needs to be local
		m.Delta = &val
		ma = append(ma, m)
	}

	if ac.BulkUpdate {
		resp, err := sendMetrics(ctx, ac, ma)
		if err == nil {
			resp.Body.Close()
			//reset counter after successful transefer
			for k := range ma {
				metric := ma[k]
				if metric.MType == "counter" {
					m.Counters[metric.ID] = 0
				}
			}
		}
	} else {
		for k := range ma {
			metric := ma[k]

			resp, err := sendMetric(ctx, ac, &metric)
			if err == nil {
				resp.Body.Close()
				//reset counter after successful transefer
				if metric.MType == "counter" {
					m.Counters[metric.ID] = 0
				}
			}
		}
	}
}

// rate limited payload sender goroutine
func payloadSender(ctx context.Context, ac app.AgentConfig, wg *sync.WaitGroup, id int, jobs <-chan uuid.UUID) {
	defer wg.Done()

	logger.Debug(fmt.Sprintf("rated-sender(%d): init", id))

	for uid := range jobs {
		logger.Info(fmt.Sprintf("rated-sender(%d): send metrics [%s]", id, uid))

		sendPayload(ctx, ac, ms)

		// no results needed in current configuration, in future can return err for specific UUID
		//results <- nil
	}

	logger.Debug(fmt.Sprintf("rated-sender(%d): channel closed", id))
}

func sendMetric(ctx context.Context, ac app.AgentConfig, metric *domain.Metrics) (*http.Response, error) {
	var resp *http.Response
	var err error

	switch ac.APIVersion {
	case "v1":
		var vs string

		switch metric.MType {
		case "gauge":
			vs = fmt.Sprintf("%f", *metric.Value)
		case "counter":
			vs = fmt.Sprintf("%d", *metric.Delta)
		default:
			logger.Error(fmt.Sprintf("sendMetric ERROR: unsupported metric type [%s]", metric.MType))
		}

		resp, err = PostValueV1(ctx, ac, metric.MType, metric.ID, vs)
	case "v2":
		jsonres, jsonerr := json.Marshal(metric)
		if jsonerr != nil {
			logger.Error(fmt.Sprintf("sendMetric ERROR: JSON marshaling failed [%s]", jsonerr))
			return nil, jsonerr
		}

		buf := bytes.NewBuffer(jsonres)

		logger.Debug(fmt.Sprintf("sendMetric: POST body %s", buf))

		resp, err = PostValueV2(ctx, ac, buf)
	default:
		logger.Error(fmt.Sprintf("sendMetric ERROR: unsupported API version %s", ac.APIVersion))
	}

	if err != nil {
		logger.Error(fmt.Sprintf("sendMetric ERROR posting value: %s, %s", metric.ID, err))
		return nil, err
	}
	if resp.StatusCode != 200 {
		logger.Info(fmt.Sprintf("response Code: %v", resp.StatusCode))
		logger.Info(fmt.Sprintf("response Status: %v", resp.Status))
		logger.Info(fmt.Sprintf("response Headers: %v", resp.Header))
	}
	return resp, err
}

func sendMetrics(ctx context.Context, ac app.AgentConfig, ma []domain.Metrics) (*http.Response, error) {
	var resp *http.Response

	jsonres, jsonerr := json.Marshal(ma)
	if jsonerr != nil {
		logger.Error(fmt.Sprintf("sendMetrics ERROR: JSON marshaling failed [%s]", jsonerr))
		return nil, jsonerr
	}

	buf := bytes.NewBuffer(jsonres)

	logger.Debug(fmt.Sprintf("sendMetrics: POST body %s", buf))

	backoff := func(ctx context.Context) error {
		//var err error

		bresp, err := PostValueV2(ctx, ac, buf)
		resp = bresp //handle linter bug, does not see body closure. set //nolint:bodyerror in prod environment
		if err == nil {
			bresp.Body.Close() //handle linter bug, does not see body closure. set //nolint:bodyerror in prod environment
		}

		return app.HandleRetriableWeb(err, "error sending data")
	}

	err := app.DoRetry(ctx, ac.MaxConnectionRetries, backoff)
	if err != nil {
		logger.Error(fmt.Sprintf("ERROR bulk posting, %s", err))
		return nil, err
	}

	if resp.StatusCode != 200 {
		logger.Info(fmt.Sprintf("response Code: %v", resp.StatusCode))
		logger.Info(fmt.Sprintf("response Status: %v", resp.Status))
		logger.Info(fmt.Sprintf("response Headers: %v", resp.Header))
	}
	return resp, nil
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
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c
	logger.Info("agent: received ^C - shutting down")

	// tell the goroutines to stop
	logger.Info("agent: telling goroutines to stop")
	cancel()

	// and wait for them to reply back
	wg.Wait()
	logger.Info("agent: shutdown")
}

func agent(ctx context.Context, wg *sync.WaitGroup) {
	//execute to exit wait group
	defer wg.Done()

	// statically linked variables (YP iter20 requirement)
	logger.Info(fmt.Sprintf("Build version: %s", naIfEmpty(buildVersion)))
	logger.Info(fmt.Sprintf("Build date: %s", naIfEmpty(buildDate)))
	logger.Info(fmt.Sprintf("Build commit: %s", naIfEmpty(buildCommit)))

	logger.Info(fmt.Sprintf("agent: using endpoint %s", ac.Endpoint))
	logger.Info(fmt.Sprintf("agent: poll interval %d", ac.PollInterval))
	logger.Info(fmt.Sprintf("agent: report interval %d", ac.ReportInterval))
	logger.Info(fmt.Sprintf("agent: encryption=%v", cryptor.CanEncrypt()))
	logger.Info(fmt.Sprintf("agent: signed messaging=%v", signer.UseSignedMessaging()))
	logger.Info(fmt.Sprintf("agent: rate limit=%v", ac.RateLimit))

	//add optional rate limiter as required by technical specs
	if ac.UseRateLimit {
		sendJobs = make(chan uuid.UUID, ac.RateLimit)
		for w := 1; w <= int(ac.RateLimit); w++ {
			wg.Add(1)
			go payloadSender(ctx, ac, wg, w, sendJobs)
		}
	}

	wg.Add(1)
	go collector(ctx, ac, wg)
	wg.Add(1)
	go collectorPs(ctx, ac, wg)
	wg.Add(1)
	go reporter(ctx, ac, wg)

	<-ctx.Done()
	logger.Info("agent: shutdown requested")

	// close channel on exit
	if ac.UseRateLimit {
		close(sendJobs)
	}

	// // shut down gracefully with timeout of 5 seconds max
	// _, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	logger.Info("agent: stopped")
}

func naIfEmpty(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

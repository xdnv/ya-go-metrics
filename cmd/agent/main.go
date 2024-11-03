// the main agent module provides agent (metric sender) function
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"internal/adapters/cryptor"
	"internal/adapters/logger"
	"internal/adapters/retrier"
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

// store agent IP address to send in header
var agentIP string

// universal Message data structure for both HTTP and gPRC communication
type Message struct {
	Address     string
	ContentType string
	Body        *bytes.Buffer
	Metadata    map[string]string
}

// NewMessage constructor for Message
func NewMessage() *Message {
	return &Message{
		Body:     new(bytes.Buffer),       // init Body as a new bytes.Buffer
		Metadata: make(map[string]string), // init the map
	}
}

// universal Response data structure for both HTTP and gPRC communication
type Response struct {
	StatusCode    int
	Status        string
	ContentLength int64
	Body          *bytes.Buffer
	Metadata      map[string][]string
}

// NewResponse constructor for Response
func NewResponse() *Response {
	return &Response{
		Body:     new(bytes.Buffer),         // init Body as a new bytes.Buffer
		Metadata: make(map[string][]string), // init the map
	}
}

// converts the http.Response object to Response
func NewHTTPResponse(r *http.Response) (*Response, error) {
	res := &Response{
		StatusCode:    r.StatusCode,
		Status:        r.Status,
		ContentLength: r.ContentLength,
		Body:          new(bytes.Buffer),         // init Body as a new bytes.Buffer
		Metadata:      make(map[string][]string), // init the map
	}

	//read out body
	if _, err := io.Copy(res.Body, r.Body); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	//read out Header copying all the slices
	for key, values := range r.Header {
		res.Metadata[key] = append([]string(nil), values...)
	}

	return res, nil
}

// simple HTTP post function
func PostData(address string, contentType string, body *bytes.Buffer) (*http.Response, error) {
	resp, err := http.Post(address, contentType, body)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// simple HTTP post function based on Message input format
func PostMessage(m *Message) (*Response, error) {
	resp, err := http.Post(m.Address, m.ContentType, m.Body)
	if err != nil {
		return nil, err
	}

	res, err := NewHTTPResponse(resp)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// extended HTTP post function based on Message input format, supports headers & more
func PostMessageExtended(m *Message) (*Response, error) {

	r, err := http.NewRequest("POST", m.Address, m.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	r.Close = true //whether to close the connection after replying to this request (for servers) or after sending the request (for clients).

	// set HTTP headers
	for k, v := range m.Metadata {
		r.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}

	res, err := NewHTTPResponse(resp)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// HTTP post metric value using API v1
func PostValueV1(ctx context.Context, ac app.AgentConfig, counterType string, counterName string, value string) (*Response, error) {
	m := NewMessage()
	m.Address = fmt.Sprintf("http://%s/update/%s/%s/%s", ac.Endpoint, counterType, counterName, value)
	m.ContentType = "text/plain"
	m.Body = nil

	switch ac.TransportMode {
	case domain.TRANSPORT_HTTP:
		return PostMessage(m)
	case domain.TRANSPORT_GRPC:
		return PostGRPC(ctx, ac, counterType, counterName, value)
	}
	return nil, errors.New("unsupported transport mode: " + ac.TransportMode)
}

// HTTP post metric value using API v2
func PostValueV2(ctx context.Context, ac app.AgentConfig, body *bytes.Buffer) (*Response, error) {
	m := NewMessage()
	m.ContentType = "application/json"
	m.Body = body

	m.Address = fmt.Sprintf("http://%s/updates/", ac.Endpoint)
	if !ac.BulkUpdate {
		m.Address = fmt.Sprintf("http://%s/update/", ac.Endpoint)
	}

	// //older API
	// if !ac.UseCompression {
	// 	return PostMessage(m)
	// }

	// set metadata for extended posting
	m.Metadata["Content-Type"] = m.ContentType
	m.Metadata["Accept-Encoding"] = "gzip"

	// set real client IP
	if agentIP != "" {
		m.Metadata["X-Real-IP"] = agentIP
	}

	//optionally encrypt message
	_, err := encryptMessage(m)
	if err != nil {
		return nil, err
	}

	//compress message
	if ac.UseCompression {
		err = compressMessage(m)
		if err != nil {
			return nil, err
		}
	}

	// for security reasons, body has to be signed after compression & encryption
	signMessage(m)

	switch ac.TransportMode {
	case domain.TRANSPORT_HTTP:
		return PostMessageExtended(m)
	case domain.TRANSPORT_GRPC:
		return PostMessageGRPC(m, ac.BulkUpdate)
	}
	return nil, errors.New("unsupported transport mode: " + ac.TransportMode)
}

func signMessage(m *Message) error {
	if !signer.IsSignedMessagingEnabled() {
		return nil
	}

	sig, err := signer.GetSignature(m.Body.Bytes())
	if err != nil {
		return err
	}

	m.Metadata[signer.GetSignatureToken()] = base64.URLEncoding.EncodeToString(sig)

	return nil
}

func encryptMessage(m *Message) (bool, error) {
	if !cryptor.CanEncrypt() {
		return false, nil
	}

	msg, err := cryptor.Encrypt(m.Body.Bytes())
	if err != nil {
		return false, err
	}

	m.Body.Reset()             // Clear the buffer
	_, err = m.Body.Write(msg) // Write encrypted data back to buffer
	if err != nil {
		return false, err
	}

	m.Metadata["X-Encrypted"] = "true"

	return true, nil
}

func compressMessage(m *Message) error {
	var buf bytes.Buffer

	g := gzip.NewWriter(&buf)
	if _, err := g.Write(m.Body.Bytes()); err != nil {
		return err
	}
	if err := g.Close(); err != nil {
		return err
	}

	m.Body.Reset()                      // Clear the buffer
	_, err := m.Body.Write(buf.Bytes()) // Write compressed data back to buffer
	if err != nil {
		return err
	}

	m.Metadata["Content-Encoding"] = "gzip"

	return nil
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
		_, err := sendMetrics(ctx, ac, ma)
		if err == nil {
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

			_, err := sendMetric(ctx, ac, &metric)
			if err == nil {
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

	logger.Debugf("rated-sender(%d): init", id)

	for uid := range jobs {
		logger.Infof("rated-sender(%d): send metrics [%s]", id, uid)

		sendPayload(ctx, ac, ms)

		// no results needed in current configuration, in future can return err for specific UUID
		//results <- nil
	}

	logger.Debugf("rated-sender(%d): channel closed", id)
}

func sendMetric(ctx context.Context, ac app.AgentConfig, metric *domain.Metrics) (*Response, error) {
	var resp *Response
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
			logger.Errorf("sendMetric ERROR: unsupported metric type [%s]", metric.MType)
		}

		resp, err = PostValueV1(ctx, ac, metric.MType, metric.ID, vs)
	case "v2":
		jsonres, jsonerr := json.Marshal(metric)
		if jsonerr != nil {
			logger.Errorf("sendMetric ERROR: JSON marshaling failed [%s]", jsonerr)
			return nil, jsonerr
		}

		buf := bytes.NewBuffer(jsonres)

		logger.Debugf("sendMetric: POST body %s", buf)

		resp, err = PostValueV2(ctx, ac, buf)
	default:
		logger.Errorf("sendMetric ERROR: unsupported API version %s", ac.APIVersion)
	}

	if err != nil {
		logger.Errorf("sendMetric ERROR posting value: %s, %s", metric.ID, err)
		return nil, err
	}
	//TODO: move or remove to stop using "resp" in this function
	// switch to local Result structure with the same fields + body as bytes.Buffer
	if resp.StatusCode != 200 {
		logger.Infof("response Code: %v", resp.StatusCode)
		logger.Infof("response Status: %v", resp.Status)
		logger.Infof("response Headers: %v", resp.Metadata)
	}
	return resp, err
}

func sendMetrics(ctx context.Context, ac app.AgentConfig, ma []domain.Metrics) (*Response, error) {
	var resp *Response

	jsonres, jsonerr := json.Marshal(ma)
	if jsonerr != nil {
		logger.Errorf("sendMetrics ERROR: JSON marshaling failed [%s]", jsonerr)
		return nil, jsonerr
	}

	buf := bytes.NewBuffer(jsonres)

	logger.Debugf("sendMetrics: POST body %s", buf)

	backoff := func(ctx context.Context) error {
		var err error

		resp, err = PostValueV2(ctx, ac, buf)

		return retrier.HandleRetriableWeb(err, "error sending data")
	}

	err := retrier.DoRetry(ctx, ac.MaxConnectionRetries, backoff)
	if err != nil {
		logger.Errorf("ERROR bulk posting, %s", err)
		return nil, err
	}

	//TODO: move or remove to stop using "resp" in this function
	// switch to local Result structure with the same fields + body as bytes.Buffer
	if resp.StatusCode != 200 {
		logger.Infof("response Code: %v", resp.StatusCode)
		logger.Infof("response Status: %v", resp.Status)
		logger.Infof("response Headers: %v", resp.Metadata)
	}
	return resp, nil
}

// returns first non-loopback local IP address
func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	var localIP string
	for _, addr := range addrs {
		// Check if address is IPv4 and not loopback
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			localIP = ipnet.IP.String()
			break
		}
	}

	if localIP == "" {
		return "", errors.New("no valid local IP address found")
	}

	return localIP, nil
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
	logger.Infof("Build version: %s", naIfEmpty(buildVersion))
	logger.Infof("Build date: %s", naIfEmpty(buildDate))
	logger.Infof("Build commit: %s", naIfEmpty(buildCommit))

	logger.Infof("agent: transport mode %s", ac.TransportMode)
	logger.Infof("agent: using endpoint %s", ac.Endpoint)
	logger.Infof("agent: poll interval %d", ac.PollInterval)
	logger.Infof("agent: report interval %d", ac.ReportInterval)
	logger.Infof("agent: encryption=%v", cryptor.CanEncrypt())
	logger.Infof("agent: signed messaging=%v", signer.IsSignedMessagingEnabled())
	logger.Infof("agent: rate limit=%v", ac.RateLimit)

	//iter24: send local IP to server
	localIP, err := getLocalIP()
	if err != nil {
		logger.Errorf("error getting local IP-address, %s", err.Error())
	}
	agentIP = localIP

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

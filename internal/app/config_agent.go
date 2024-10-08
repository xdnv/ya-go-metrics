// agent configuration module provides app-wide configuration structure with easy init
package app

import (
	"flag"
	"os"
	"strconv"

	"internal/adapters/signer"
)

// agent configuration
type AgentConfig struct {
	Endpoint             string // the address:port server endpoint to send metric data
	ReportInterval       int64  // metric reporting frequency in seconds
	PollInterval         int64  // metric poll interval in seconds
	LogLevel             string // log verbosity (log level)
	APIVersion           string // API version to send metric data. Recent is v2
	UseCompression       bool   // activate gzip compression
	BulkUpdate           bool   // activate bulk JSON metric update
	MaxConnectionRetries uint64 // Connection retries for retriable functions (does not include original request. 0 to disable)
	UseRateLimit         bool   // flag option to enable or disable rate limiter
	RateLimit            int64  // max simultaneous connections to server (rate limit)
}

// set agent configuration using command line arguments and/or environment variables
func InitAgentConfig() AgentConfig {
	var MsgKey string

	cf := AgentConfig{}

	cf.APIVersion = "v2"        // activate JSON support
	cf.UseCompression = true    // activate gzip compression
	cf.BulkUpdate = true        // activate bulk JSON metric update
	cf.MaxConnectionRetries = 3 // Connection retries for retriable functions (does not include original request. 0 to disable)

	//set defaults and read command line
	flag.StringVar(&cf.Endpoint, "a", "localhost:8080", "the address:port server endpoint to send metric data")
	flag.Int64Var(&cf.PollInterval, "p", 2, "metric poll interval in seconds")
	flag.Int64Var(&cf.ReportInterval, "r", 10, "metric reporting frequency in seconds")
	flag.Int64Var(&cf.RateLimit, "l", 0, "max simultaneous connections to server, set 0 to disable rate limit")
	flag.StringVar(&MsgKey, "k", "", "key to use signed messaging, empty value disables signing")
	flag.StringVar(&cf.LogLevel, "v", "info", "log verbosity (log level)")
	flag.Parse()

	//parse env variables
	if val, found := os.LookupEnv("ADDRESS"); found {
		cf.Endpoint = val
	}
	if val, found := os.LookupEnv("POLL_INTERVAL"); found {
		intval, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			cf.PollInterval = intval
		}
	}
	if val, found := os.LookupEnv("REPORT_INTERVAL"); found {
		intval, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			cf.ReportInterval = intval
		}
	}
	if val, found := os.LookupEnv("RATE_LIMIT"); found {
		intval, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			cf.RateLimit = intval
		}
	}
	if val, found := os.LookupEnv("KEY"); found {
		MsgKey = val
	}
	if val, found := os.LookupEnv("LOG_LEVEL"); found {
		cf.LogLevel = val
	}

	if cf.Endpoint == "" {
		panic("PANIC: endpoint address:port is not set")
	}
	if cf.PollInterval == 0 {
		panic("PANIC: poll interval is not set")
	}
	if cf.ReportInterval == 0 {
		panic("PANIC: report interval is not set")
	}
	if cf.LogLevel == "" {
		panic("PANIC: log level is not set")
	}

	//set signing mode
	signer.SetKey(MsgKey)

	cf.UseRateLimit = (cf.RateLimit > 0)

	return cf
}

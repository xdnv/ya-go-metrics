package app

import (
	"flag"
	"os"
	"strconv"
)

// agent config storage
type AgentConfig struct {
	Endpoint             string
	ReportInterval       int64
	PollInterval         int64
	LogLevel             string
	APIVersion           string
	UseCompression       bool
	BulkUpdate           bool
	MaxConnectionRetries uint64
	UseSignedMessaging   bool
	MsgKey               string
}

func InitAgentConfig() AgentConfig {
	cf := AgentConfig{}

	//activate JSON support
	cf.APIVersion = "v2"
	//activate gzip compression
	cf.UseCompression = true
	//activate bulk JSON metric update
	cf.BulkUpdate = true
	//Connection retries for retriable functions (does not include original request. 0 to disable)
	cf.MaxConnectionRetries = 3

	//set defaults and read command line
	flag.StringVar(&cf.Endpoint, "a", "localhost:8080", "the address:port server endpoint to send metric data")
	flag.Int64Var(&cf.PollInterval, "p", 2, "metric poll interval in seconds")
	flag.Int64Var(&cf.ReportInterval, "r", 10, "metric reporting frequency in seconds")
	flag.StringVar(&cf.MsgKey, "k", "", "key to use signed messaging, empty value disables signing")
	flag.StringVar(&cf.LogLevel, "l", "info", "log level")
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
	if val, found := os.LookupEnv("KEY"); found {
		cf.MsgKey = val
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
	cf.UseSignedMessaging = (cf.MsgKey != "")

	return cf
}

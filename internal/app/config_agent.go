// agent configuration module provides app-wide configuration structure with easy init
package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"internal/adapters/cryptor"
	"internal/adapters/signer"
	"internal/domain"
)

// agent configuration
type AgentConfig struct {
	TransportMode        string `json:"transport_mode,omitempty"`    // data exchange transport mode: http or grpc
	Endpoint             string `json:"address,omitempty"`           // the address:port server endpoint to send metric data
	ReportInterval       int64  `json:"report_interval,omitempty"`   // metric reporting frequency in seconds
	PollInterval         int64  `json:"poll_interval,omitempty"`     // metric poll interval in seconds
	LogLevel             string `json:"log_level,omitempty"`         // log verbosity (log level)
	APIVersion           string `json:""`                            // API version to send metric data. Recent is v2
	UseCompression       bool   `json:""`                            // activate gzip compression
	BulkUpdate           bool   `json:""`                            // activate bulk JSON metric update
	MaxConnectionRetries uint64 `json:""`                            // Connection retries for retriable functions (does not include original request. 0 to disable)
	UseRateLimit         bool   `json:""`                            // flag option to enable or disable rate limiter
	RateLimit            int64  `json:"rate_limit,omitempty"`        // max simultaneous connections to server (rate limit)
	MessageSignature     string `json:"message_signature,omitempty"` // key to use signed messaging, empty value disables signing
	CryptoKeyPath        string `json:"crypto_key,omitempty"`        // path to public crypto key (to encrypt messages to server)
	ConfigFilePath       string `json:""`                            //path to JSON config file
}

// NewConfig initializes a Config with default values
func NewAgentConfig() AgentConfig {
	return AgentConfig{
		ConfigFilePath:   "",
		TransportMode:    domain.TRANSPORT_HTTP,
		Endpoint:         domain.ENDPOINT,
		PollInterval:     2,
		ReportInterval:   10,
		RateLimit:        0,
		MessageSignature: "",
		CryptoKeyPath:    "",
		LogLevel:         domain.LOGLEVEL,
	}
}

// custom command line parser to read config file name before flag.Parse() -- iter22 requirement
func ParseAgentConfigFile(cf *AgentConfig) {
	for i, arg := range os.Args {
		if arg == "-config" {
			if i+1 < len(os.Args) {
				cf.ConfigFilePath = strings.TrimSpace(os.Args[i+1])
			}
		}
	}
	if val, found := os.LookupEnv("CONFIG"); found {
		cf.ConfigFilePath = val
	}

	if cf.ConfigFilePath == "" {
		return
	}

	jcf := NewAgentConfig()

	file, err := os.Open(cf.ConfigFilePath)
	if err != nil {
		panic(fmt.Sprintf("PANIC: error reading config file: %s", err.Error()))
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&jcf); err != nil {
		panic(fmt.Sprintf("PANIC: error decoding JSON config: %s", err.Error()))
	}

	cf.TransportMode = jcf.TransportMode
	cf.Endpoint = jcf.Endpoint
	cf.PollInterval = jcf.PollInterval
	cf.ReportInterval = jcf.ReportInterval
	cf.RateLimit = jcf.RateLimit
	cf.MessageSignature = jcf.MessageSignature
	cf.CryptoKeyPath = jcf.CryptoKeyPath
	cf.LogLevel = jcf.LogLevel
}

// set agent configuration using command line arguments and/or environment variables
func InitAgentConfig() AgentConfig {

	cf := NewAgentConfig()
	cf.APIVersion = "v2"        // activate JSON support
	cf.UseCompression = false   // activate gzip compression //TODO: return to TRUE after debug will be finished
	cf.BulkUpdate = true        // activate bulk JSON metric update
	cf.MaxConnectionRetries = 3 // Connection retries for retriable functions (does not include original request. 0 to disable)
	cf.ConfigFilePath = ""

	//load config from command line or env variable with lowest priority
	ParseAgentConfigFile(&cf)

	//set defaults and read command line
	flag.StringVar(&cf.ConfigFilePath, "config", cf.ConfigFilePath, "path to configuration file in JSON format") //used to pass Parse() check
	flag.StringVar(&cf.TransportMode, "transport", cf.TransportMode, "data exchange transport mode: http or grpc")
	flag.StringVar(&cf.Endpoint, "a", cf.Endpoint, "the address:port server endpoint to send metric data")
	flag.Int64Var(&cf.PollInterval, "p", cf.PollInterval, "metric poll interval in seconds")
	flag.Int64Var(&cf.ReportInterval, "r", cf.ReportInterval, "metric reporting frequency in seconds")
	flag.Int64Var(&cf.RateLimit, "l", cf.RateLimit, "max simultaneous connections to server, set 0 to disable rate limit")
	flag.StringVar(&cf.MessageSignature, "k", cf.MessageSignature, "key to use signed messaging, empty value disables signing")
	flag.StringVar(&cf.CryptoKeyPath, "crypto-key", cf.CryptoKeyPath, "path to public crypto key")
	flag.StringVar(&cf.LogLevel, "v", cf.LogLevel, "log verbosity (log level)")
	flag.Parse()

	//parse env variables
	if val, found := os.LookupEnv("TRANSPORT_MODE"); found {
		cf.TransportMode = val
	}
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
		cf.MessageSignature = val
	}
	if val, found := os.LookupEnv("CRYPTO_KEY"); found {
		cf.CryptoKeyPath = val
	}
	if val, found := os.LookupEnv("LOG_LEVEL"); found {
		cf.LogLevel = val
	}

	// check for critical missing config entries
	if cf.TransportMode != domain.TRANSPORT_HTTP && cf.TransportMode != domain.TRANSPORT_GRPC {
		panic("PANIC: application transport mode set incorrectly")
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

	// set signing mode
	signer.SetKey(cf.MessageSignature)
	cf.MessageSignature = "" //for security reasons

	//set encryption logic
	cf.CryptoKeyPath = strings.TrimSpace(cf.CryptoKeyPath)
	if cf.CryptoKeyPath != "" {
		err := cryptor.LoadPublicKey(cf.CryptoKeyPath)
		if err != nil {
			panic("PANIC: failed to load crypto key " + err.Error())
		}
		cryptor.EnableEncryption(true)
	}

	// rate limiter global en\disable
	cf.UseRateLimit = (cf.RateLimit > 0)

	return cf
}

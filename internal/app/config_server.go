// server configuration module provides app-wide configuration structure with easy init
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
)

// defines main session storage type based on server config given
type StorageType int

// session storage type
const (
	Memory StorageType = iota
	File
	Database
)

// return session storage type as string value
func (t StorageType) String() string {
	switch t {
	case Memory:
		return "Memory"
	case File:
		return "File"
	case Database:
		return "Database"
	}
	return fmt.Sprintf("Unknown (%d)", t)
}

// server configuration
type ServerConfig struct {
	Endpoint                 string      `json:"address,omitempty"`           // the address:port endpoint for server to listen
	StoreInterval            int64       `json:"store_interval,omitempty"`    // interval in seconds to store metrics in datafile, set 0 for synchronous output
	StorageMode              StorageType `json:""`                            // session storage type
	MaxConnectionRetries     uint64      `json:""`                            // max connection retries to storage objects
	FileStoragePath          string      `json:"file_storage_path,omitempty"` // full datafile path to store/load state of metrics. empty value shuts off metric dumps
	RestoreMetrics           bool        `json:"restore_metrics,omitempty"`   // load metrics from datafile on server start, boolean
	DatabaseDSN              string      `json:"database_dsn,omitempty"`      // database DSN (format: 'host=<host> [port=port] user=<user> password=<xxxx> dbname=<mydb> sslmode=disable')
	LogLevel                 string      `json:"log_level,omitempty"`         // log level
	CompressReplies          bool        `json:"compress_replies,omitempty"`  // compress server replies, boolean
	CompressibleContentTypes []string    `json:""`                            // array of compressible mime types
	MessageSignature         string      `json:"message_signature,omitempty"` // key to use signed messaging, empty value disables signing
	CryptoKeyPath            string      `json:"crypto_key,omitempty"`        // path to private crypto key (to decrypt messages from client)
	ConfigFilePath           string      `json:""`                            //path to JSON config file
}

// NewConfig initializes a Config with default values
func NewServerConfig() ServerConfig {
	return ServerConfig{
		ConfigFilePath:   "",
		Endpoint:         "localhost:8080",
		StoreInterval:    300,
		DatabaseDSN:      "",
		MessageSignature: "",
		CryptoKeyPath:    "",
		FileStoragePath:  "/tmp/metrics-db.json",
		RestoreMetrics:   true,
		CompressReplies:  true,
		LogLevel:         "info",
	}
}

// custom command line parser to read config file name before flag.Parse() -- iter22 requirement
func ParseServerConfigFile(cf *ServerConfig) {
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

	jcf := NewServerConfig()

	file, err := os.Open(cf.ConfigFilePath)
	if err != nil {
		panic(fmt.Sprintf("PANIC: error reading config file: %s", err.Error()))
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&jcf); err != nil {
		panic(fmt.Sprintf("PANIC: error decoding JSON config: %s", err.Error()))
	}

	cf.Endpoint = jcf.Endpoint
	cf.StoreInterval = jcf.StoreInterval
	cf.DatabaseDSN = jcf.DatabaseDSN
	cf.MessageSignature = jcf.MessageSignature
	cf.CryptoKeyPath = jcf.CryptoKeyPath
	cf.FileStoragePath = jcf.FileStoragePath
	cf.RestoreMetrics = jcf.RestoreMetrics
	cf.CompressReplies = jcf.CompressReplies
	cf.LogLevel = jcf.LogLevel
}

func InitServerConfig() ServerConfig {
	cf := NewServerConfig()

	cf.CompressibleContentTypes = []string{
		"text/html",
		"application/json",
	}

	//load config from command line or env variable with lowest priority
	ParseServerConfigFile(&cf)

	flag.StringVar(&cf.ConfigFilePath, "config", cf.ConfigFilePath, "path to configuration file in JSON format") //used to pass Parse() check
	flag.StringVar(&cf.Endpoint, "a", cf.Endpoint, "the address:port endpoint for server to listen")
	flag.Int64Var(&cf.StoreInterval, "i", cf.StoreInterval, "interval in seconds to store metrics in datafile, set 0 for synchronous output")
	flag.StringVar(&cf.DatabaseDSN, "d", cf.DatabaseDSN, "database DSN (format: 'host=<host> [port=port] user=<user> password=<xxxx> dbname=<mydb> sslmode=disable')")
	flag.StringVar(&cf.MessageSignature, "k", cf.MessageSignature, "key to use signed messaging, empty value disables signing")
	flag.StringVar(&cf.CryptoKeyPath, "crypto-key", cf.CryptoKeyPath, "path to private crypto key")
	flag.StringVar(&cf.FileStoragePath, "f", cf.FileStoragePath, "full datafile path to store/load state of metrics. empty value shuts off metric dumps")
	flag.BoolVar(&cf.RestoreMetrics, "r", cf.RestoreMetrics, "load metrics from datafile on server start, boolean")
	flag.BoolVar(&cf.CompressReplies, "c", cf.CompressReplies, "compress server replies, boolean")
	flag.StringVar(&cf.LogLevel, "l", cf.LogLevel, "log level")
	flag.Parse()

	if val, found := os.LookupEnv("ADDRESS"); found {
		cf.Endpoint = val
	}
	if val, found := os.LookupEnv("STORE_INTERVAL"); found {
		intval, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			cf.StoreInterval = intval
		}
	}
	if val, found := os.LookupEnv("FILE_STORAGE_PATH"); found {
		cf.FileStoragePath = val
	}
	if val, found := os.LookupEnv("RESTORE"); found {
		boolval, err := strconv.ParseBool(val)
		if err == nil {
			cf.CompressReplies = boolval
		}
	}
	if val, found := os.LookupEnv("COMPRESS_REPLIES"); found {
		boolval, err := strconv.ParseBool(val)
		if err == nil {
			cf.RestoreMetrics = boolval
		}
	}
	if val, found := os.LookupEnv("DATABASE_DSN"); found {
		cf.DatabaseDSN = val
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
	if cf.Endpoint == "" {
		panic("PANIC: endpoint address:port is not set")
	}
	if cf.LogLevel == "" {
		panic("PANIC: log level is not set")
	}

	//set main storage type for current session
	if cf.DatabaseDSN != "" {
		cf.StorageMode = Database
	} else if cf.FileStoragePath != "" {
		cf.StorageMode = File
	} else {
		cf.StorageMode = Memory
	}

	//set signing mode
	signer.SetKey(cf.MessageSignature)
	cf.MessageSignature = "" //for security reasons

	//set encryption logic
	cf.CryptoKeyPath = strings.TrimSpace(cf.CryptoKeyPath)
	if cf.CryptoKeyPath != "" {
		err := cryptor.LoadPrivateKey(cf.CryptoKeyPath)
		if err != nil {
			panic("PANIC: failed to load crypto key " + err.Error())
		}
		cryptor.EnableEncryption(true)
	}

	return cf
}

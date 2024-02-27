package main

import (
	"flag"
	"os"
	"strconv"
)

type ServerConfig struct {
	Endpoint        string
	StoreInterval   int64
	FileStoragePath string
	RestoreMetrics  bool
	LogLevel        string
}

func InitServerConfig() ServerConfig {
	cf := ServerConfig{}

	flag.StringVar(&cf.Endpoint, "a", "localhost:8080", "the address:port endpoint for server to listen")
	flag.Int64Var(&cf.StoreInterval, "i", 300, "interval in seconds to store metrics in datafile, set 0 for synchronous output")
	flag.StringVar(&cf.FileStoragePath, "f", "/tmp/metrics-db.json", "full datafile path to store/load state of metrics. empty value shuts off metric dumps")
	flag.BoolVar(&cf.RestoreMetrics, "r", true, "load metrics from datafile on server start, boolean")
	flag.StringVar(&cf.LogLevel, "l", "info", "log level")
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
			cf.RestoreMetrics = boolval
		}
	}
	if val, found := os.LookupEnv("LOG_LEVEL"); found {
		cf.LogLevel = val
	}

	if cf.Endpoint == "" {
		panic("PANIC: endpoint address:port is not set")
	}
	if cf.LogLevel == "" {
		panic("PANIC: log level is not set")
	}

	return cf
}

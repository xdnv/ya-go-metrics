package main

import (
	"flag"
	"os"
	"strconv"
)


// agent config storage
type AgentConfig struct {
	Endpoint       string
	ReportInterval int64
	PollInterval   int64
}

func InitAgentConfig() AgentConfig {

	cf := AgentConfig{}

	//set defaults and read command line
	flag.StringVar(&cf.Endpoint, "a", "localhost:8080", "the address:port server endpoint to send metric data")
	flag.Int64Var(&cf.ReportInterval, "r", 10, "metric reporting frequency in seconds")
	flag.Int64Var(&cf.PollInterval, "p", 2, "metric poll interval in seconds")
	flag.Parse()

	//parse env variables
	if val, found := os.LookupEnv("ADDRESS"); found && (val != "") {
		cf.Endpoint = val
	}
	if val, found := os.LookupEnv("REPORT_INTERVAL"); found && (val != "") {
		intval, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			cf.ReportInterval = intval
		}
	}
	if val, found := os.LookupEnv("POLL_INTERVAL"); found && (val != "") {
		intval, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			cf.PollInterval = intval
		}
	}

	// // Access and print non-flag arguments
	// args := flag.Args()
	// fmt.Println("Non-flag arguments:", args)

	return cf
}

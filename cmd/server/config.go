package main

import (
	"flag"
	"os"
)

type ServerConfig struct {
	Endpoint string
}

func InitServerConfig() ServerConfig {

	cf := ServerConfig{}

	flag.StringVar(&cf.Endpoint, "a", "localhost:8080", "the address:port endpoint for server to listen")
	flag.Parse()

	if val, found := os.LookupEnv("ADDRESS"); found && (val != "") {
		cf.Endpoint = val
	}

	return cf
}

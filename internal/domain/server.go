package domain

import "fmt"

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
	TransportMode            string      `json:"transport_mode,omitempty"`    // data exchange transport mode: http or grpc
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
	TrustedSubnet            string      `json:"trusted_subnet,omitempty"`    // trusted agent subnet in CIDR form. use empty value to disable security check.
	ConfigFilePath           string      `json:""`                            //path to JSON config file
}

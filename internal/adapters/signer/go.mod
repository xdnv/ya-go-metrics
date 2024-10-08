module signer

go 1.22.7

require internal/adapters/logger v1.0.0

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)

replace internal/adapters/logger => ../../adapters/logger

module storage

go 1.22.7

require github.com/jackc/pgx/v5 v5.7.1

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.18.0 // indirect
)

require internal/adapters/logger v1.0.0

replace internal/adapters/logger => ../../adapters/logger

require (
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438
	github.com/sethvargo/go-retry v0.3.0
	internal/domain v1.0.0
)

replace internal/domain => ../../domain

module storage

go 1.21.5

require github.com/jackc/pgx/v5 v5.5.5

require (
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/sethvargo/go-retry v0.2.4 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

require internal/adapters/logger v1.0.0

replace internal/adapters/logger => ../../adapters/logger

require internal/app v1.0.0

replace internal/app => ../../app

require internal/domain v1.0.0

replace internal/domain => ../../domain

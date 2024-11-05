module app

go 1.22.7

require (
	internal/adapters/cryptor v1.0.0
	internal/adapters/firewall v1.0.0
	internal/adapters/signer v1.0.0
	internal/domain v1.0.0
	internal/ports/storage v1.0.0
)

require (
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.27.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	internal/adapters/logger v1.0.0 // indirect
)

replace internal/adapters/cryptor => ../adapters/cryptor

replace internal/adapters/firewall => ../adapters/firewall

replace internal/adapters/logger => ../adapters/logger

replace internal/adapters/signer => ../adapters/signer

replace internal/domain => ../domain

replace internal/ports/storage => ../ports/storage

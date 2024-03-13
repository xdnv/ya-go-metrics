module github.com/xdnv/ya-go-metrics

go 1.21.5

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/stretchr/testify v1.8.4
	go.uber.org/zap v1.27.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require internal/adapters/logger v1.0.0

replace internal/adapters/logger => ./internal/adapters/logger

require internal/app v1.0.0

replace internal/app => ./internal/app

require internal/domain v1.0.0

replace internal/domain => ./internal/domain

require internal/ports/storage v1.0.0

replace internal/ports/storage => ./internal/ports/storage

require github.com/jackc/pgx/v5 v5.5.5

module github.com/xdnv/ya-go-metrics

go 1.22.7

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/stretchr/testify v1.9.0
	go.uber.org/zap v1.27.0 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require internal/adapters/cryptor v1.0.0

replace internal/adapters/cryptor => ./internal/adapters/cryptor

require internal/adapters/firewall v1.0.0

replace internal/adapters/firewall => ./internal/adapters/firewall

require internal/adapters/logger v1.0.0

replace internal/adapters/logger => ./internal/adapters/logger

require internal/adapters/retrier v1.0.0

replace internal/adapters/retrier => ./internal/adapters/retrier

require internal/adapters/signer v1.0.0

replace internal/adapters/signer => ./internal/adapters/signer

require internal/app v1.0.0

replace internal/app => ./internal/app

require internal/domain v1.0.0

replace internal/domain => ./internal/domain

require internal/service v1.0.0

replace internal/service => ./internal/service

require internal/transport/grpc_server v1.0.0

replace internal/transport/grpc_server => ./internal/transport/grpc_server

require (
	github.com/google/uuid v1.6.0
	github.com/shirou/gopsutil/v3 v3.24.3
	golang.org/x/tools v0.26.0
	google.golang.org/grpc v1.67.1
	honnef.co/go/tools v0.5.1
	internal/ports/storage v1.0.0
)

replace internal/ports/storage => ./internal/ports/storage

require (
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.1 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	golang.org/x/crypto v0.28.0 // indirect
	golang.org/x/exp/typeparams v0.0.0-20231108232855-2478ac86f678 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
)

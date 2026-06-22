# Kratos v2 Microservices Template

A small, opinionated Go microservices scaffold built on Kratos v2: clean layering,
uniform module wiring with Uber FX, service discovery, and distributed tracing.
All services are **gRPC-only**.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Framework | Kratos v2 |
| Dependency Injection | Uber FX |
| ORM | GORM Gen (type-safe generated queries) |
| Database | PostgreSQL |
| Service Discovery | Consul |
| Config | Consul (prod) / local YAML, via Kratos config |
| Tracing | OpenTelemetry → Jaeger (OTLP) |
| Logging | zap |
| Protobuf | buf |

## Services & Ports

| Service | gRPC | Notes |
|---------|------|-------|
| auth | 9081 | JWT login / refresh / validate (calls user over gRPC; owns no DB) |
| user | 9082 | User CRUD; owner of the `users` table |
| PostgreSQL | 5432 | Database |
| Consul | 8500 | Service discovery + config (UI at :8500) |
| Jaeger | 16686 | Tracing UI (OTLP collector on 4317) |

## Quick Start

```bash
# Build all services
make build

# Run a single service locally against configs/<svc>.yaml
make run-auth
make run-user

# Or bring up the full stack (Postgres + Consul + Jaeger + services)
cd deploy && docker compose up -d
docker compose ps          # wait until healthy

# Consul UI:  http://localhost:8500
# Jaeger UI:  http://localhost:16686
```

On startup, Consul auto-loads each service's config from `deploy/configs/<svc>/config.yaml`
(see `deploy/init-consul-config.sh`).

## Project Structure

```
.
├── api/<svc>/v1/            # Generated gRPC contract (DO NOT edit by hand)
├── proto/<svc>/v1/          # Proto sources — change the interface here
├── app/<svc>/               # Self-contained service (can be split into its own repo)
│   ├── cmd/<svc>/main.go    # Entry point: flag + bootstrap.Run + layer Modules
│   └── internal/
│       ├── conf/            # Service-private config proto + generated code
│       ├── server/          # gRPC server, middleware & route registration
│       ├── service/         # Implements the pb XxxServer; DTO ↔ domain mapping
│       ├── biz/             # Use cases + Repo interfaces + domain types/errors
│       └── data/            # Repo impls: GORM Gen (owns a table) or a gRPC client to another service
├── pkg/                     # Cross-service, business-agnostic infrastructure
│   ├── bootstrap/           # Generic startup wiring (config, log, tracer, kratos app)
│   ├── log/                 # zap logger + kratos/gorm adapters
│   ├── registry/            # Consul registrar / discovery
│   ├── auth/                # JWT manager
│   ├── conf/                # Shared config proto (service / registry / log)
│   └── model/               # Shared GORM data models
├── configs/<svc>.yaml       # Local run config (per service)
├── deploy/                  # Dockerfile, docker-compose, init scripts, prod configs
└── tools/gen/               # GORM Gen configuration
```

## Architecture

Each service follows a single-direction four-layer flow with dependency inversion —
`biz` defines `Repo` interfaces, `data` implements them; `biz` never imports `data`:

```
server ──▶ service ──▶ biz ◀── data
 transport   adapter    domain   storage
```

Wiring is done with Uber FX: one `fx.go` per layer exposing a `Module` named
`<svc>.<layer>`, assembled in `cmd/<svc>/main.go` via `bootstrap.Run[conf.Bootstrap]`.

The `data` layer implements `biz.Repo` against whatever backs the data — its own
database, or **another service over gRPC**. Each table has a single owning service;
others reach it through that owner's API rather than sharing the table. The auth
service owns no table: its `AuthUserRepo` is a Kratos gRPC client to the user
service (endpoint `discovery:///user` via Consul, or a direct `host:port` locally).
bcrypt verification stays in the user service behind the internal `VerifyCredentials`
RPC, so password hashes never cross the wire.

## Configuration

Config is **independent per service** — each service loads only its own:

- **Local:** `configs/<svc>.yaml` (used by `make run-<svc>`).
- **Prod:** Consul, key prefix `config/<svc>/`.
- **Priority:** env vars > Consul > local file.

Each YAML splits into a **common section** (`service` / `registry` / `log`, defined
once in `pkg/conf`) and a **service-private section** (`server` / `data` and anything
service-specific such as auth's `jwt_secret`, defined in `app/<svc>/internal/conf`).

Environment variables (highest priority) override config values:

| Env | Overrides | Read by |
|-----|-----------|---------|
| `DB_DSN` | database connection string | data layer of table-owning services (user) |
| `USER_SERVICE_ENDPOINT` | user-service gRPC target | auth data layer |
| `JWT_SECRET` | auth token signing secret | auth service |
| `CONSUL_ADDR` | Consul address (config source + registry) | bootstrap |
| `CONSUL_CONFIG_PATH` | Consul key prefix | bootstrap |
| `SERVICE_NAME` / `SERVICE_VERSION` | service identity | bootstrap |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | tracing endpoint; unset disables tracing | bootstrap |

## Common Commands

```bash
# codegen (after changing proto or models)
make generate       # proto + gorm
make proto          # buf generate
make gorm           # GORM Gen query code

# develop
make run-<svc>      # run one service locally (e.g. make run-auth)
make test           # tests (race + cover)
make fmt            # gofmt
make lint           # golangci-lint
make tidy           # go mod tidy

# build
make build          # all services into bin/
make build-<svc>    # one service
make clean          # remove bin/

# deploy stack (Postgres + Consul + Jaeger + services)
make up             # build & start
make logs           # follow logs
make down           # stop
```

Running `make` with no target prints the full list.

## Adding a Service

1. Define the interface in `proto/<svc>/v1/<svc>.proto`, then `make proto`.
2. Create `app/<svc>/internal/{conf,biz,data,service,server}`, one `fx.go` per layer.
3. Write `app/<svc>/cmd/<svc>/main.go` wiring the layer Modules via `bootstrap.Run`.
4. Add `configs/<svc>.yaml` + `deploy/` config, and append `<svc>` to `SERVICES` in the Makefile.

## Requirements

- Go 1.25+
- Docker & Docker Compose
- buf (protobuf tooling)
- golangci-lint (optional, for linting)

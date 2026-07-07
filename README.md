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
| Token Store | Redis |
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
| Redis | 6379 | JWT denylist + refresh token rotation state |
| Consul | 8500 | Service discovery + config (UI at :8500) |
| Jaeger | 16686 | Tracing UI (OTLP collector on 4317) |

## Quick Start

```bash
# Build all services
make build

# Local services need PostgreSQL and Redis running first.
# Then apply migrations, start user, and start auth.
make migrate-up
make run-user
make run-auth

# Or bring up the full stack
make up
docker compose -f deploy/docker-compose.yml ps

# Consul UI:  http://localhost:8500
# Jaeger UI:  http://localhost:16686
```

On startup, `consul-init` loads each service's config from
`deploy/configs/<svc>/config.yaml` into Consul KV before services start.

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
│   ├── middleware/          # shared authn / validation middleware
│   ├── registry/            # Consul registrar / discovery
│   ├── auth/                # JWT manager
│   ├── conf/                # Shared config proto (registry / log)
│   └── model/               # Shared GORM data models
├── migrations/              # golang-migrate SQL migrations
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
bcrypt verification stays in the user service behind the internal
`VerifyCredentials` RPC, so password hashes never cross the wire.

## Authentication

Auth issues an access token and a refresh token. Access tokens are short-lived;
refresh tokens are stored in Redis, rotated on every refresh, and consumed with
`GETDEL`. Reusing an old refresh token revokes all refresh tokens for that user.
`Logout` deny-lists the access token JTI until expiry and consumes the refresh
token when one is provided.

The user service protects RPCs with JWT middleware:

| RPC | Accepted token type |
|-----|---------------------|
| `VerifyCredentials` | `service` only |
| `GetUser` | `access` or `service` |
| `CreateUser` / `ListUsers` | `access` |
| `UpdateUser` / `DeleteUser` | `access`, owner only (`claims.UserID` must match the target id) |

### Revocation trust boundary

The user service validates access tokens **locally** (signature check only) and does
not consult the Redis denylist — a deliberate trade-off: stateless verification, no
per-RPC Redis round-trip, and token state stays owned by auth alone. A logged-out
access token therefore keeps working against the user service until it expires; the
revocation delay is bounded by the access-token TTL (default 15 min). Callers that
need immediate revocation must go through auth's `Validate` (e.g. from an API gateway).

### No public registration

`CreateUser` itself requires an access token, so there is no unauthenticated sign-up
path. The first user comes from the demo seed migration (`0002`, demo-only); replace
it with your own provisioning flow in production.

This template uses one shared HS256 platform secret read by both auth and user
services. Generate one with:

```bash
openssl rand -base64 32
```

For production, prefer an asymmetric key flow such as RS256/EdDSA public-key
distribution.

## Configuration

Config is **independent per service** — each service loads only its own:

- **Local:** `configs/<svc>.yaml` (used by `make run-<svc>`).
- **Prod:** Consul, key prefix `config/<svc>/`.
- **Priority:** env vars > Consul > local file.

Each YAML splits into a **common section** (`registry` / `log`, defined once in
`pkg/conf`) and a **service-private section** (`server` / `data` and anything
service-specific such as auth's `jwt_secret`, defined in `app/<svc>/internal/conf`).

Service **name** and **version** are not configured — name is the literal passed to
`bootstrap.Run` in `main.go`, version is stamped at build time via `-ldflags` (run
`make build`; local runs report `dev`). Check a binary with `<svc> -version`.

Environment variables (highest priority) override config values:

| Env | Overrides | Read by |
|-----|-----------|---------|
| `DB_DSN` | database connection string | data layer of table-owning services (user) |
| `REDIS_ADDR` | Redis address | auth data layer |
| `USER_SERVICE_ENDPOINT` | user-service gRPC target | auth data layer |
| `JWT_SECRET` | token signing secret; must be 32+ bytes | auth + user |
| `CONSUL_ADDR` | Consul address (config source + registry) | bootstrap |
| `CONSUL_CONFIG_PATH` | Consul key prefix | bootstrap |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | tracing endpoint; unset disables tracing | bootstrap |
| `OTEL_TRACES_SAMPLER_ARG` | trace sample ratio, default `1.0` | bootstrap |
| `OTEL_EXPORTER_OTLP_INSECURE` | OTLP gRPC insecure mode, default `true` | bootstrap |

## Migrations

Schema is managed by golang-migrate SQL files under `migrations/`. The Docker
stack runs migrations automatically before `user` starts. For local development:

```bash
make migrate-up
make migrate-down
```

## Common Commands

```bash
# codegen (after changing proto or models)
make generate       # proto + gorm
make proto          # buf generate
make gorm           # GORM Gen query code

# develop
make run-<svc>      # run one service locally (e.g. make run-auth)
make test           # tests (race + cover)
make fmt            # golangci-lint fmt
make lint           # golangci-lint
make buf-lint       # buf lint
make check          # fmt + lint + buf-lint + test
make tidy           # go mod tidy
make migrate-up     # apply DB migrations
make migrate-down   # roll back one migration

# build
make build          # all services into bin/
make build-<svc>    # one service
make clean          # remove bin/

# deploy stack (Postgres + Redis + Consul + Jaeger + services)
make up             # build & start
make logs           # follow logs
make down           # stop
```

Running `make` with no target prints the full list.

## Adding a Service

1. Define the interface in `proto/<svc>/v1/<svc>.proto`, then `make proto`.
2. Create `app/<svc>/internal/{conf,biz,data,service,server}`, one `fx.go` per layer.
3. Write `app/<svc>/cmd/<svc>/main.go` wiring the layer Modules via `bootstrap.Run`.
4. Add `configs/<svc>.yaml` + `deploy/configs/<svc>/config.yaml`.
5. Add Docker Compose service wiring if the service belongs in the local stack.

## Requirements

- Go 1.25+
- Docker & Docker Compose
- buf (protobuf tooling)
- golangci-lint (optional, for linting)

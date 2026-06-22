# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

A small, opinionated **gRPC-only** Kratos v2 microservices scaffold. `README.md` has the
full architecture, port map, env-var table, and "Adding a Service" steps — read it for the
big picture. This file covers the conventions that span multiple files and bite if missed.

## Commands

```bash
make run-<svc>          # run one service locally vs configs/<svc>.yaml (e.g. make run-auth)
make build              # all services → bin/ ; make build-<svc> for one
make test               # go test -race -cover ./...
make lint               # golangci-lint (v2; gofumpt+goimports formatters; api/ excluded)
make fmt                # gofmt -s -w .
make generate           # regenerate codegen: proto + gorm (see below)
make up / logs / down   # full docker stack (Postgres + Consul + Jaeger + services)
```

Run a single test: `go test -race -run '^TestName$' ./app/<svc>/internal/<layer>/`.

## Codegen — generated files, never hand-edit

- `api/<svc>/v1/*.pb.go` + `*_grpc.pb.go` and each service's `internal/conf/conf.pb.go`
  come from `make proto` (`buf generate`). Edit the `.proto` source, not the output:
  service contracts in `proto/<svc>/v1/`, private config in `app/<svc>/internal/conf/conf.proto`.
- `app/<svc>/internal/data/query/*.gen.go` is GORM Gen output from `make gorm`
  (`go run tools/gen/main.go`), driven by structs in `pkg/model`. To generate a new table:
  add the model to `pkg/model`, then add `g.ApplyBasic(model.X{})` and the query path to
  `targets` in `tools/gen/main.go`.
- Run `make generate` after touching any proto or model.

## FX wiring conventions

Wiring is Uber FX, not Kratos wire. Each layer has one `fx.go` exposing
`var Module = fx.Module("<svc>.<layer>", …)`. `cmd/<svc>/main.go` is the only assembly point:

```go
bootstrap.Run[conf.Bootstrap](flagConf, "config/<svc>/",
    bootstrap.WithKratosApp(), data.Module, biz.Module, service.Module, server.Module)
```

- `bootstrap.Run` loads config, inits zap + OTEL tracer, and supplies into the FX graph:
  `kratosconfig.Config`, `*conf.Bootstrap` (service-private), `*conf.CommonConfig` (shared),
  `*zap.Logger`, plus registry/discovery and service identity via **named** tags
  (`name:"service_id"`, etc. — see `pkg/bootstrap/providers.go`).
- The **server** Module annotates its gRPC server into `group:"servers"` as a
  `transport.Server`; `NewKratosApp` collects that group. A new transport must join the group.
- Resource constructors return `(*T, func(), error)`; the `func()` cleanup is registered onto
  the FX lifecycle via a small `fx.Invoke(registerLifecycle)` in that layer's `fx.go`.
- Cross-cutting params travel in `fx.In`/`fx.Out` structs (see `GRPCServerParams`, `AppParams`).

## Layering & dependency inversion

`server → service → biz ← data` (transport → adapter → domain → storage).

- `biz` owns domain types, errors, and the `Repo` **interface**; `data` implements it.
  `biz` must never import `data`. Each impl asserts `var _ biz.XxxRepo = (*xxxRepo)(nil)`.
- `service` implements the generated `pb.XxxServer` and maps DTO ↔ biz domain only.
- A `data` Repo is backed by **either** GORM (the service owns that table) **or** a gRPC client
  to another service. One owning service per table — others reach it through that owner's API,
  never by sharing the table. `auth` owns no DB: its repo is a gRPC client to `user`, and
  bcrypt verification stays behind user's `VerifyCredentials` RPC so hashes never cross the wire.

## Config

Independent per service. Local `configs/<svc>.yaml`; prod Consul prefix `config/<svc>/`.
Priority: **env > Consul > local**. Common section (`service`/`registry`/`log`) is defined in
`pkg/conf`; the private section (`server`/`data`/service-specific like `jwt_secret`) in
`app/<svc>/internal/conf`. Env overrides are applied inline at the read site via
`cmp.Or(os.Getenv("X"), cfg…)` — grep for the var names in `README.md`'s env table.

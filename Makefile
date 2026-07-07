.DEFAULT_GOAL := help

BUILD_DIR := bin
COMPOSE   := docker compose -f deploy/docker-compose.yml
MIGRATE_IMAGE ?= migrate/migrate:v4.18.3

VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -ldflags "-s -w -X kratos-template/pkg/bootstrap.Version=$(VERSION)"
export VERSION   # so docker-compose build args pick it up via `make up`

.PHONY: help generate proto gorm build clean test lint fmt tidy buf-lint buf-breaking check migrate-up migrate-down up down logs

## help: list available targets
help:
	@grep -hE '^## ' $(MAKEFILE_LIST) | sed 's/## //' | awk -F': ' '{printf "  %-14s %s\n", $$1, $$2}'

# codegen (run after changing proto or models)

## generate: regenerate everything (proto + gorm)
generate: proto gorm

## proto: generate gRPC/protobuf code via buf
proto:
	buf generate

## gorm: generate type-safe GORM query code
gorm:
	go run tools/gen/main.go

# develop

## run-<svc>: run one service locally (e.g. make run-auth)
run-%:
	go run ./app/$*/cmd/$*

## test: run all tests (race + coverage)
test:
	go test -race -cover ./...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## fmt: format all Go code
fmt:
	golangci-lint fmt

## buf-lint: lint protobuf sources
buf-lint:
	buf lint

## buf-breaking: check protobuf breaking changes against main
buf-breaking:
	buf breaking --against '.git#branch=main'

## check: run local quality checks
check: fmt lint buf-lint buf-breaking test

## tidy: tidy go modules
tidy:
	go mod tidy

# build

## build: build all services into bin/
build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/ ./app/*/cmd/*

## build-<svc>: build one service (e.g. make build-auth)
build-%:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$* ./app/$*/cmd/$*

## clean: remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## migrate-up: apply database migrations
migrate-up:
	docker run --rm -v $(PWD)/migrations:/migrations:ro $(MIGRATE_IMAGE) -path=/migrations -database "$${DB_DSN:-postgres://postgres:postgres@host.docker.internal:5432/user_db?sslmode=disable}" up

## migrate-down: roll back one database migration
migrate-down:
	docker run --rm -v $(PWD)/migrations:/migrations:ro $(MIGRATE_IMAGE) -path=/migrations -database "$${DB_DSN:-postgres://postgres:postgres@host.docker.internal:5432/user_db?sslmode=disable}" down 1

# deploy stack (Postgres + Redis + Consul + Jaeger + services)

## up: build and start the full stack
up:
	$(COMPOSE) up -d --build

## down: stop the stack
down:
	$(COMPOSE) down

## logs: follow stack logs
logs:
	$(COMPOSE) logs -f

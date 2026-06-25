.DEFAULT_GOAL := help

BUILD_DIR := bin
COMPOSE   := docker compose -f deploy/docker-compose.yml

VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS   := -ldflags "-s -w -X kratos-template/pkg/bootstrap.Version=$(VERSION)"
export VERSION   # so docker-compose build args pick it up via `make up`

.PHONY: help generate proto gorm build clean test lint fmt tidy up down logs

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
	gofmt -s -w .

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

# deploy stack (Postgres + Consul + Jaeger + services)

## up: build and start the full stack
up:
	$(COMPOSE) up -d --build

## down: stop the stack
down:
	$(COMPOSE) down

## logs: follow stack logs
logs:
	$(COMPOSE) logs -f

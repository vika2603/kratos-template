.PHONY: all build clean test generate proto docker help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Binary names
SERVICES=auth user asset gateway

# Build directory
BUILD_DIR=bin

# Docker
DOCKER_COMPOSE=docker-compose -f deploy/docker-compose.yml

all: generate build

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all        - Generate code and build all services"
	@echo "  build      - Build all services"
	@echo "  clean      - Clean build artifacts"
	@echo "  test       - Run tests"
	@echo "  generate   - Generate proto and GORM code"
	@echo "  proto      - Generate protobuf code using buf"
	@echo "  gorm       - Generate GORM query code"
	@echo "  docker     - Build Docker images"
	@echo "  docker-up  - Start all services with Docker Compose"
	@echo "  docker-down- Stop all services"
	@echo "  fmt        - Format Go code"
	@echo "  lint       - Run linter"
	@echo "  tidy       - Tidy Go modules"

## build: Build all services
build:
	@echo "Building services..."
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		if [ "$$svc" = "gateway" ]; then \
			$(GOBUILD) -o $(BUILD_DIR)/$$svc ./cmd/$$svc; \
		else \
			$(GOBUILD) -o $(BUILD_DIR)/$$svc ./cmd/$$svc; \
		fi; \
	done
	@echo "Build complete!"

## build-%: Build specific service (e.g., make build-auth)
build-%:
	@echo "Building $*..."
	@if [ "$*" = "gateway" ]; then \
		$(GOBUILD) -o $(BUILD_DIR)/$* ./cmd/$*; \
	else \
		$(GOBUILD) -o $(BUILD_DIR)/$* ./cmd/$*; \
	fi

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)/*
	@echo "Clean complete!"

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./...

## test-%: Run tests for specific service (e.g., make test-auth)
test-%:
	@echo "Testing $*..."
	$(GOTEST) -v -race -cover ./app/$*/...

## generate: Generate all code (proto + gorm)
generate: proto gorm

## gen-gateway: Generate gateway code with hz
gen-gateway:
	@echo "Generating gateway code..."
	@rm -rf app/gateway/biz
	@cd app/gateway && for proto in idl/auth.proto idl/asset.proto idl/user.proto; do \
		hz update --idl $$proto --module gateway --out_dir . \
			--customize_package template/package.yaml; \
	done
	@$(GOFMT) -w app/gateway/
	@cd app/gateway && go build ./...

## proto: Generate protobuf code using buf
proto:
	@echo "Generating protobuf code..."
	buf generate
	@echo "Proto generation complete!"

## gorm: Generate GORM query code
gorm:
	@echo "Generating GORM code..."
	$(GOCMD) run tools/gen/main.go
	@echo "GORM generation complete!"

## docker: Build Docker images for all services
docker:
	@echo "Building Docker images..."
	@for svc in $(SERVICES); do \
		echo "Building $$svc image..."; \
		docker build -t kratos-template-$$svc:latest -f app/$$svc/Dockerfile .; \
	done
	@echo "Docker build complete!"

## docker-up: Start all services with Docker Compose
docker-up:
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d
	@echo "Services started!"

## docker-down: Stop all services
docker-down:
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down
	@echo "Services stopped!"

## docker-logs: View logs for all services
docker-logs:
	$(DOCKER_COMPOSE) logs -f

## docker-logs-%: View logs for specific service
docker-logs-%:
	$(DOCKER_COMPOSE) logs -f $*

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	@echo "Format complete!"

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...
	@echo "Lint complete!"

## tidy: Tidy Go modules
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy
	@echo "Tidy complete!"

## init-db: Initialize database
init-db:
	@echo "Initializing database..."
	$(DOCKER_COMPOSE) exec -T postgres psql -U postgres -f /docker-entrypoint-initdb.d/init-db.sql
	@echo "Database initialized!"

## run-%: Run specific service locally (e.g., make run-auth)
run-%:
	@echo "Running $*..."
	@if [ "$*" = "gateway" ]; then \
		$(GOCMD) run ./cmd/$* -conf ./app/$*/configs/config.yaml; \
	else \
		$(GOCMD) run ./cmd/$* -conf ./app/$*/configs/config.yaml; \
	fi

# Kratos v2 Microservices Template

Production-ready Go microservices scaffold built with Kratos v2, featuring dependency injection, service discovery, and distributed tracing.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Framework | Kratos v2 |
| Dependency Injection | Uber FX |
| ORM | GORM Gen |
| Database | PostgreSQL |
| Service Discovery | Consul |
| Tracing | OpenTelemetry + Jaeger |
| Metrics | Prometheus + Grafana |
| Protocol Buffers | buf |

## Port Planning

| Service | HTTP | gRPC | Notes |
|---------|------|------|-------|
| Gateway | 8080 | - | Public entry point |
| Auth | 8081 | 9081 | Internal |
| User | 8082 | 9082 | Internal |
| Consul | 8500 | - | Service discovery |
| Jaeger | 16686 | - | Tracing UI |
| PostgreSQL | 5432 | - | Database |

## Quick Start

```bash
# Start infrastructure and all services
cd deploy && docker-compose up -d

# Wait for services to be healthy (30-60 seconds)
docker-compose ps

# Test the API
curl http://localhost:8080/healthz

# Access web interfaces
# Consul UI: http://localhost:8500
# Jaeger UI: http://localhost:16686
```

## Project Structure

```
.
├── api/                    # Protocol buffer definitions
│   ├── auth/v1/           # Auth service API
│   ├── gateway/v1/        # Gateway API
│   └── user/v1/           # User service API
├── app/                    # Service implementations
│   ├── auth/              # Authentication service
│   ├── gateway/           # API gateway (BFF)
│   ├── user/              # User management service
├── pkg/                    # Shared libraries
│   ├── bootstrap/         # Service initialization
│   ├── consul/            # Service discovery client
│   ├── errors/            # Error handling
│   ├── middleware/        # HTTP/gRPC middleware
│   └── model/             # Shared data models
├── deploy/                 # Deployment configurations
│   ├── docker-compose.yml # Local development stack
│   ├── init-db.sql        # Database initialization
└── tools/                  # Code generation tools
    └── gen/               # GORM Gen configuration
```

## Documentation

- [Architecture](docs/architecture.md) - System design and service interactions
- [Development Guide](docs/development.md) - Local development workflow
- [Deployment Guide](docs/deployment.md) - Production deployment instructions

## Key Features

- **Microservices Architecture**: Three independent services with clear boundaries
- **Dependency Injection**: Uber FX for clean, testable code structure
- **Type-Safe ORM**: GORM Gen for compile-time query validation
- **Service Discovery**: Consul for dynamic service registration and health checks
- **Distributed Tracing**: OpenTelemetry integration with Jaeger backend
- **API Gateway**: BFF pattern with centralized routing and authentication
- **Observability**: Full metrics, logs, and traces with Prometheus and Grafana

## Common Commands

```bash
# Build all services
make build

# Run tests
make test

# Generate proto and GORM code
make generate

# Run specific service locally
make run-auth

# View service logs
make docker-logs-auth

# Stop all services
make docker-down
```

## Requirements

- Go 1.21+
- Docker & Docker Compose
- buf (Protocol buffer tooling)
- golangci-lint (optional, for linting)

## License

MIT

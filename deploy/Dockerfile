FROM golang:1.25-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download


COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/gateway ./cmd/gateway && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/auth ./cmd/auth && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/user ./cmd/user && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/asset ./cmd/asset

FROM alpine:3.19 AS base
RUN apk --no-cache add ca-certificates tzdata

FROM base AS gateway
WORKDIR /app
COPY --from=builder /out/gateway .
EXPOSE 8080
ENTRYPOINT ["/app/gateway"]

FROM base AS auth
WORKDIR /app
COPY --from=builder /out/auth .
EXPOSE 9081
ENTRYPOINT ["/app/auth"]

FROM base AS user
WORKDIR /app
COPY --from=builder /out/user .
EXPOSE 9082
ENTRYPOINT ["/app/user"]

FROM base AS asset
WORKDIR /app
COPY --from=builder /out/asset .
EXPOSE 9083
ENTRYPOINT ["/app/asset"]

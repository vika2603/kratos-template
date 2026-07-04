FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/grpc-ecosystem/grpc-health-probe@latest


COPY . .

ARG VERSION=dev
ENV LDFLAGS="-s -w -X kratos-template/pkg/bootstrap.Version=$VERSION"

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="$LDFLAGS" -o /out/auth ./app/auth/cmd/auth && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="$LDFLAGS" -o /out/user ./app/user/cmd/user

FROM alpine:3.21 AS base
RUN apk --no-cache add ca-certificates tzdata && adduser -D -u 10001 app
USER app

FROM base AS auth
WORKDIR /app
COPY --from=builder /out/auth .
COPY --from=builder /go/bin/grpc-health-probe /bin/grpc_health_probe
EXPOSE 9081
HEALTHCHECK CMD ["/bin/grpc_health_probe", "-addr=:9081"]
ENTRYPOINT ["/app/auth"]

FROM base AS user
WORKDIR /app
COPY --from=builder /out/user .
COPY --from=builder /go/bin/grpc-health-probe /bin/grpc_health_probe
EXPOSE 9082
HEALTHCHECK CMD ["/bin/grpc_health_probe", "-addr=:9082"]
ENTRYPOINT ["/app/user"]

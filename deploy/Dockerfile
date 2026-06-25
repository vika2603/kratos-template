FROM golang:1.25-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download


COPY . .

ARG VERSION=dev
ENV LDFLAGS="-s -w -X kratos-template/pkg/bootstrap.Version=$VERSION"

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="$LDFLAGS" -o /out/auth ./app/auth/cmd/auth && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="$LDFLAGS" -o /out/user ./app/user/cmd/user

FROM alpine:3.19 AS base
RUN apk --no-cache add ca-certificates tzdata

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


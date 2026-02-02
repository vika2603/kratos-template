FROM golang:1.25.5-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o asset ./app/asset/cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/asset .
COPY --from=builder /build/app/asset/configs /app/configs

EXPOSE 8083 9083

ENTRYPOINT ["/app/asset"]
CMD ["-conf", "/app/configs/config.yaml"]

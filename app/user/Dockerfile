FROM golang:1.25.5-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user ./app/user/cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/user .
COPY --from=builder /build/app/user/configs /app/configs

EXPOSE 8082 9082

ENTRYPOINT ["/app/user"]
CMD ["-conf", "/app/configs/config.yaml"]

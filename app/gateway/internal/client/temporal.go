package client

import (
	"os"

	"go.temporal.io/sdk/client"

	"kratos-template/app/gateway/internal/conf"
)

func NewTemporalClient(cfg *conf.Bootstrap) (client.Client, error) {
	hostPort := os.Getenv("TEMPORAL_ADDR")
	if hostPort == "" {
		hostPort = cfg.Temporal.HostPort
	}
	return client.Dial(client.Options{
		HostPort:  hostPort,
		Namespace: cfg.Temporal.Namespace,
	})
}

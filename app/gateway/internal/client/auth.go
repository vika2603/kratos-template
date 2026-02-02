package client

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	authv1 "kratos-template/api/auth/v1"
	"kratos-template/app/gateway/internal/conf"
)

func NewAuthClient(cfg *conf.Bootstrap, logger log.Logger, reg registry.Discovery) (authv1.AuthServiceClient, error) {
	timeout, err := time.ParseDuration(cfg.Client.Auth.Timeout)
	if err != nil {
		timeout = 5 * time.Second
	}

	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+cfg.Client.Auth.DiscoveryName),
		grpc.WithDiscovery(reg),
		grpc.WithTimeout(timeout),
		grpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
			logging.Client(logger),
		),
	)
	if err != nil {
		return nil, err
	}

	return authv1.NewAuthServiceClient(conn), nil
}

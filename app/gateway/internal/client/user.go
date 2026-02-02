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

	userv1 "kratos-template/api/user/v1"
	"kratos-template/app/gateway/internal/conf"
)

func NewUserClient(cfg *conf.Bootstrap, logger log.Logger, reg registry.Discovery) (userv1.UserServiceClient, error) {
	timeout, err := time.ParseDuration(cfg.Client.User.Timeout)
	if err != nil {
		timeout = 5 * time.Second
	}

	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+cfg.Client.User.DiscoveryName),
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

	return userv1.NewUserServiceClient(conn), nil
}

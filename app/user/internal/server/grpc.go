package server

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.uber.org/fx"

	v1 "kratos-template/api/user/v1"
	"kratos-template/app/user/internal/conf"
)

type GRPCServerParams struct {
	fx.In
	Config      *conf.Bootstrap
	Logger      log.Logger
	UserService v1.UserServiceServer
}

func NewGRPCServer(params GRPCServerParams) (*grpc.Server, error) {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(params.Logger),
		),
	}

	if params.Config.Server.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(params.Config.Server.Grpc.Addr))
	}

	if params.Config.Server.Grpc.Timeout != "" {
		timeout, err := time.ParseDuration(params.Config.Server.Grpc.Timeout)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.Timeout(timeout))
	}

	srv := grpc.NewServer(opts...)
	v1.RegisterUserServiceServer(srv, params.UserService)

	return srv, nil
}

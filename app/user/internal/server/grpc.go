package server

import (
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	v1 "kratos-template/api/user/v1"
	"kratos-template/app/user/internal/conf"
	"kratos-template/pkg/log/adapter"
)

type GRPCServerParams struct {
	fx.In
	Config      *conf.Bootstrap
	Logger      *zap.Logger
	UserService v1.UserServiceServer
}

func NewGRPCServer(params GRPCServerParams) (*grpc.Server, error) {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(adapter.NewKratosAdapter(params.Logger)),
		),
	}

	if params.Config.Server.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(params.Config.Server.Grpc.Addr))
	}

	if t := params.Config.Server.Grpc.GetTimeout(); t != nil {
		opts = append(opts, grpc.Timeout(t.AsDuration()))
	}

	srv := grpc.NewServer(opts...)
	v1.RegisterUserServiceServer(srv, params.UserService)

	return srv, nil
}

package server

import (
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	v1 "kratos-template/api/auth/v1"
	"kratos-template/app/auth/internal/conf"
	"kratos-template/pkg/log/adapter"
)

type GRPCServerParams struct {
	fx.In
	Config      *conf.Bootstrap
	Logger      *zap.Logger
	AuthService v1.AuthServiceServer
}

func NewGRPCServer(params GRPCServerParams) (*grpc.Server, error) {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(adapter.NewKratosAdapter(params.Logger)),
		),
	}

	if addr := params.Config.GetServer().GetGrpc().GetAddr(); addr != "" {
		opts = append(opts, grpc.Address(addr))
	}

	if t := params.Config.GetServer().GetGrpc().GetTimeout(); t != nil {
		opts = append(opts, grpc.Timeout(t.AsDuration()))
	}

	srv := grpc.NewServer(opts...)
	v1.RegisterAuthServiceServer(srv, params.AuthService)

	return srv, nil
}

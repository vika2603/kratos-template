package server

import (
	"kratos-template/app/auth/internal/conf"
	"kratos-template/pkg/bootstrap"
	"time"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.uber.org/fx"
	"go.uber.org/zap"

	v1 "kratos-template/api/auth/v1"
)

type GRPCServerParams struct {
	fx.In
	Config      *conf.Bootstrap
	Logger      *zap.Logger
	AuthService v1.AuthServiceServer
}

func NewGRPCServer(params GRPCServerParams) *grpc.Server {
	grpcCfg := params.Config.GetServer().GetGrpc()
	var timeout time.Duration
	if t := grpcCfg.GetTimeout(); t != nil {
		timeout = t.AsDuration()
	}
	return bootstrap.BuildGRPCServer(
		bootstrap.GRPCServerConfig{
			Addr:    grpcCfg.GetAddr(),
			Timeout: timeout,
		},
		params.Logger,
		func(srv *grpc.Server) {
			v1.RegisterAuthServiceServer(srv, params.AuthService)
		},
	)
}

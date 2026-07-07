package server

import (
	"cmp"
	"kratos-template/app/user/internal/conf"
	"kratos-template/pkg/bootstrap"
	"kratos-template/pkg/middleware/authn"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc/health"

	v1 "kratos-template/api/user/v1"

	pkgauth "kratos-template/pkg/auth"
)

type GRPCServerParams struct {
	fx.In
	Config      *conf.Bootstrap
	Logger      *zap.Logger
	Health      *health.Server
	UserService v1.UserServiceServer
}

func NewGRPCServer(params GRPCServerParams) (*grpc.Server, error) {
	manager, err := pkgauth.NewJWTManager(
		cmp.Or(os.Getenv("JWT_SECRET"), params.Config.GetAuth().GetJwtSecret()),
		0,
		0,
	)
	if err != nil {
		return nil, err
	}

	serviceOnly := selector.Server(authn.Server(manager, pkgauth.TokenTypeService)).
		Path(v1.UserService_VerifyCredentials_FullMethodName).
		Build()
	accessOrService := selector.Server(authn.Server(manager, pkgauth.TokenTypeAccess, pkgauth.TokenTypeService)).
		Path(v1.UserService_GetUser_FullMethodName).
		Build()
	accessOnly := selector.Server(authn.Server(manager, pkgauth.TokenTypeAccess)).
		Path(
			v1.UserService_CreateUser_FullMethodName,
			v1.UserService_UpdateUser_FullMethodName,
			v1.UserService_DeleteUser_FullMethodName,
			v1.UserService_ListUsers_FullMethodName,
		).
		Build()

	grpcCfg := params.Config.GetServer().GetGrpc()
	var timeout time.Duration
	if t := grpcCfg.GetTimeout(); t != nil {
		timeout = t.AsDuration()
	}
	return bootstrap.BuildGRPCServer(
		bootstrap.GRPCServerConfig{
			Addr:    grpcCfg.GetAddr(),
			Timeout: timeout,
			Health:  params.Health,
		},
		params.Logger,
		func(srv *grpc.Server) {
			v1.RegisterUserServiceServer(srv, params.UserService)
		},
		serviceOnly,
		accessOrService,
		accessOnly,
	), nil
}

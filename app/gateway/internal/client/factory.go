package client

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"
	kratosgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"go.uber.org/fx"
	"google.golang.org/grpc"

	assetv1 "kratos-template/api/asset/v1"
	authv1 "kratos-template/api/auth/v1"
	userv1 "kratos-template/api/user/v1"
	"kratos-template/app/gateway/internal/conf"
)

// ClientFactory consolidates shared middleware setup for gRPC clients.
type ClientFactory struct {
	cfg    *conf.Bootstrap
	logger log.Logger
	reg    registry.Discovery
}

// NewClientFactory creates a new ClientFactory instance.
func NewClientFactory(cfg *conf.Bootstrap, logger log.Logger, reg registry.Discovery) *ClientFactory {
	return &ClientFactory{
		cfg:    cfg,
		logger: logger,
		reg:    reg,
	}
}

// newGRPCConn creates a gRPC connection with shared middleware configuration.
func (f *ClientFactory) newGRPCConn(discoveryName, timeoutStr string) (*grpc.ClientConn, error) {
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 5 * time.Second
	}

	conn, err := kratosgrpc.DialInsecure(
		context.Background(),
		kratosgrpc.WithEndpoint("discovery:///"+discoveryName),
		kratosgrpc.WithDiscovery(f.reg),
		kratosgrpc.WithTimeout(timeout),
		kratosgrpc.WithMiddleware(
			recovery.Recovery(),
			tracing.Client(),
			logging.Client(f.logger),
		),
	)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// AuthClient creates an authenticated gRPC client for the Auth service.
func (f *ClientFactory) AuthClient() (authv1.AuthServiceClient, error) {
	conn, err := f.newGRPCConn(f.cfg.Client.Auth.DiscoveryName, f.cfg.Client.Auth.Timeout)
	if err != nil {
		return nil, err
	}
	return authv1.NewAuthServiceClient(conn), nil
}

// UserClient creates an authenticated gRPC client for the User service.
func (f *ClientFactory) UserClient() (userv1.UserServiceClient, error) {
	conn, err := f.newGRPCConn(f.cfg.Client.User.DiscoveryName, f.cfg.Client.User.Timeout)
	if err != nil {
		return nil, err
	}
	return userv1.NewUserServiceClient(conn), nil
}

// AssetClient creates an authenticated gRPC client for the Asset service.
func (f *ClientFactory) AssetClient() (assetv1.AssetServiceClient, error) {
	conn, err := f.newGRPCConn(f.cfg.Client.Asset.DiscoveryName, f.cfg.Client.Asset.Timeout)
	if err != nil {
		return nil, err
	}
	return assetv1.NewAssetServiceClient(conn), nil
}

// ProvideClients returns an fx.Option that provides all gRPC clients at once.
func ProvideClients() fx.Option {
	return fx.Options(
		fx.Provide(NewClientFactory),
		fx.Provide(func(f *ClientFactory) (authv1.AuthServiceClient, error) {
			return f.AuthClient()
		}),
		fx.Provide(func(f *ClientFactory) (userv1.UserServiceClient, error) {
			return f.UserClient()
		}),
		fx.Provide(func(f *ClientFactory) (assetv1.AssetServiceClient, error) {
			return f.AssetClient()
		}),
	)
}

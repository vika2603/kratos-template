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

	assetv1 "kratos-template/api/asset/v1"
	"kratos-template/app/gateway/internal/conf"
)

func NewAssetClient(cfg *conf.Bootstrap, logger log.Logger, reg registry.Discovery) (assetv1.AssetServiceClient, error) {
	timeout, err := time.ParseDuration(cfg.Client.Asset.Timeout)
	if err != nil {
		timeout = 5 * time.Second
	}

	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("discovery:///"+cfg.Client.Asset.DiscoveryName),
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

	return assetv1.NewAssetServiceClient(conn), nil
}

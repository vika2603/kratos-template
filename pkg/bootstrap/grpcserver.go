package bootstrap

import (
	"kratos-template/pkg/log/adapter"
	"kratos-template/pkg/middleware/validate"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"go.uber.org/zap"
)

type GRPCServerConfig struct {
	Addr    string
	Timeout time.Duration
}

func BuildGRPCServer(
	cfg GRPCServerConfig,
	logger *zap.Logger,
	register func(*kgrpc.Server),
	extra ...middleware.Middleware,
) *kgrpc.Server {
	stack := []middleware.Middleware{
		recovery.Recovery(),
		tracing.Server(),
		logging.Server(adapter.NewKratosAdapter(logger)),
	}
	stack = append(stack, extra...)
	stack = append(stack, validate.Server())

	opts := []kgrpc.ServerOption{kgrpc.Middleware(stack...)}
	if cfg.Addr != "" {
		opts = append(opts, kgrpc.Address(cfg.Addr))
	}
	if cfg.Timeout > 0 {
		opts = append(opts, kgrpc.Timeout(cfg.Timeout))
	}

	// kgrpc.NewServer registers the standard grpc.health.v1 service by default.
	srv := kgrpc.NewServer(opts...)
	register(srv)
	return srv
}

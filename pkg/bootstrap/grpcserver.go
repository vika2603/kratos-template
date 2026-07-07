package bootstrap

import (
	"kratos-template/pkg/log/adapter"
	"kratos-template/pkg/middleware/validate"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/ratelimit"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel"
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
		metrics.Server(metricsOptions(logger)...),
		// BBR before tracing: shed load without paying span overhead.
		ratelimit.Server(),
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

// metricsOptions builds RED instruments; on failure the middleware degrades
// to a no-op rather than blocking startup.
func metricsOptions(logger *zap.Logger) []metrics.Option {
	meter := otel.Meter("kratos-template")
	var opts []metrics.Option

	counter, err := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	if err != nil {
		logger.Warn("metrics requests counter disabled", zap.Error(err))
	} else {
		opts = append(opts, metrics.WithRequests(counter))
	}

	seconds, err := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	if err != nil {
		logger.Warn("metrics seconds histogram disabled", zap.Error(err))
	} else {
		opts = append(opts, metrics.WithSeconds(seconds))
	}
	return opts
}

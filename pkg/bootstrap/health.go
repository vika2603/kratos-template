package bootstrap

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	healthInterval     = 10 * time.Second
	healthCheckTimeout = 3 * time.Second
)

// HealthChecker is one dependency probe; data layers contribute them via
// fx group:"health_checkers".
type HealthChecker struct {
	Name  string
	Check func(context.Context) error
}

// NewHealthServer starts NOT_SERVING; the monitor flips it once the first
// probe round passes.
func NewHealthServer() *health.Server {
	hs := health.NewServer()
	hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	return hs
}

type healthMonitorParams struct {
	fx.In
	Logger   *zap.Logger
	Checkers []HealthChecker `group:"health_checkers"`
}

func runHealthMonitor(lc fx.Lifecycle, hs *health.Server, params healthMonitorParams) {
	if len(params.Checkers) == 0 {
		hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		return
	}

	stop := make(chan struct{})
	done := make(chan struct{})
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go monitorHealth(hs, params, stop, done)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			close(stop)
			// Don't let a checker that ignores its context hang shutdown.
			select {
			case <-done:
			case <-ctx.Done():
				return ctx.Err()
			}
			hs.Shutdown()
			return nil
		},
	})
}

func monitorHealth(hs *health.Server, params healthMonitorParams, stop, done chan struct{}) {
	defer close(done)

	last := grpc_health_v1.HealthCheckResponse_NOT_SERVING
	probe := func() {
		status := grpc_health_v1.HealthCheckResponse_SERVING
		for _, c := range params.Checkers {
			ctx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
			err := c.Check(ctx)
			cancel()
			if err != nil {
				status = grpc_health_v1.HealthCheckResponse_NOT_SERVING
				params.Logger.Warn("health check failed",
					zap.String("checker", c.Name), zap.Error(err))
			}
		}
		if status != last {
			params.Logger.Info("health status changed",
				zap.String("from", last.String()), zap.String("to", status.String()))
			last = status
		}
		hs.SetServingStatus("", status)
	}

	probe()
	ticker := time.NewTicker(healthInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			probe()
		}
	}
}

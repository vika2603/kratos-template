package bootstrap

import (
	"context"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type AppParams struct {
	fx.In
	Logger     *zap.Logger
	Shutdowner fx.Shutdowner
	Registrar  registry.Registrar `optional:"true"`
	Servers    []transport.Server `group:"servers"`
	ServiceID  string             `name:"service_id"`
	Name       string             `name:"service_name"`
	Version    string             `name:"service_version"`
}

func NewKratosApp(lc fx.Lifecycle, params AppParams) *kratos.App {
	started := make(chan struct{})
	opts := []kratos.Option{
		kratos.ID(params.ServiceID),
		kratos.Name(params.Name),
		kratos.Version(params.Version),
		kratos.Logger(adapter.NewKratosAdapter(params.Logger)),
		kratos.Server(params.Servers...),
		kratos.AfterStart(func(context.Context) error {
			close(started)
			return nil
		}),
	}

	if params.Registrar != nil {
		opts = append(opts, kratos.Registrar(params.Registrar))
	}

	app := kratos.New(opts...)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			runErr := make(chan error, 1)
			go func() {
				if err := app.Run(); err != nil {
					log.Errorf("kratos app run error: %v", err)
					runErr <- err
					if shutdownErr := params.Shutdowner.Shutdown(fx.ExitCode(1)); shutdownErr != nil {
						log.Errorf("fx shutdown after kratos run error failed: %v", shutdownErr)
					}
				}
			}()
			select {
			case <-started:
				return nil
			case err := <-runErr:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		},
		OnStop: func(ctx context.Context) error {
			return app.Stop()
		},
	})

	return app
}

func WithKratosApp() fx.Option {
	return fx.Options(
		fx.Provide(NewKratosApp),
		fx.Provide(NewHealthServer),
		fx.Invoke(runHealthMonitor),
		fx.Invoke(func(*kratos.App) {}),
	)
}

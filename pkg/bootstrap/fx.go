package bootstrap

import (
	"context"

	"github.com/go-kratos/kratos/v2"
	kratosconfig "github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
)

func CommonLifecycleOptions(shutdown func(context.Context) error) fx.Option {
	return fx.Options(
		fx.Invoke(func(lc fx.Lifecycle, c kratosconfig.Config) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					return c.Close()
				},
			})
		}),
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{OnStop: shutdown})
		}),
	)
}

type AppParams struct {
	fx.In
	Logger     *zap.Logger
	Registrar  registry.Registrar `optional:"true"`
	Servers    []transport.Server `group:"servers"`
	ServiceID  string             `name:"service_id"`
	Name       string             `name:"service_name"`
	Version    string             `name:"service_version"`
	Metadata   map[string]string  `name:"service_metadata" optional:"true"`
	ShutdownFn func() error       `name:"app_shutdown" optional:"true"`
}

func NewKratosApp(lc fx.Lifecycle, params AppParams) *kratos.App {
	opts := []kratos.Option{
		kratos.ID(params.ServiceID),
		kratos.Name(params.Name),
		kratos.Version(params.Version),
		kratos.Logger(adapter.NewKratosAdapter(params.Logger)),
		kratos.Server(params.Servers...),
	}

	if params.Registrar != nil {
		opts = append(opts, kratos.Registrar(params.Registrar))
	}

	if params.Metadata != nil {
		opts = append(opts, kratos.Metadata(params.Metadata))
	}

	app := kratos.New(opts...)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := app.Run(); err != nil {
					log.Fatalf("kratos app run error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if params.ShutdownFn != nil {
				if err := params.ShutdownFn(); err != nil {
					log.Errorf("custom shutdown error: %v", err)
				}
			}
			return app.Stop()
		},
	})

	return app
}

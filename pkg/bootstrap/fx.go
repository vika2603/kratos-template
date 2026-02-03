package bootstrap

import (
	"context"

	"github.com/go-kratos/kratos/v2"
	kratoslog "github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	"go.uber.org/fx"
)

type AppParams struct {
	fx.In
	Logger     kratoslog.Logger
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
		kratos.Logger(params.Logger),
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
					kratoslog.Fatalf("kratos app run error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if params.ShutdownFn != nil {
				if err := params.ShutdownFn(); err != nil {
					kratoslog.Errorf("custom shutdown error: %v", err)
				}
			}
			return app.Stop()
		},
	})

	return app
}

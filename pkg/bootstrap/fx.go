package bootstrap

import (
	"context"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport"
	"go.uber.org/fx"
)

type AppParams struct {
	fx.In
	Logger     log.Logger
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

func ProvideServiceInfo(id, name, version string, metadata map[string]string) fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func() string { return id },
				fx.ResultTags(`name:"service_id"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() string { return name },
				fx.ResultTags(`name:"service_name"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() string { return version },
				fx.ResultTags(`name:"service_version"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() map[string]string { return metadata },
				fx.ResultTags(`name:"service_metadata"`),
			),
		),
	)
}

func AsServer(f interface{}) fx.Option {
	return fx.Provide(
		fx.Annotate(
			f,
			fx.As(new(transport.Server)),
			fx.ResultTags(`group:"servers"`),
		),
	)
}

type BootstrapConfig interface {
	GetService() ServiceConfig
	GetRegistry() RegistryConfig
	GetTracing() TracingConfig
	GetLog() LogConfig
}

type ServiceConfig interface {
	GetName() string
	GetVersion() string
}

type RegistryConfig interface {
	GetConsul() ConsulConfig
}

type ConsulConfig interface {
	GetAddress() string
	GetScheme() string
}

type TracingConfig interface {
	GetJaeger() JaegerConfig
	GetSampleRate() float64
}

type JaegerConfig interface {
	GetEndpoint() string
}

type LogConfig interface {
	GetLevel() string
	GetEnv() string
}

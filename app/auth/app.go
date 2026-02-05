package auth

import (
	"context"
	"flag"
	"os"

	"github.com/go-kratos/kratos/v2"
	kratosconfig "github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"kratos-template/app/auth/internal/biz"
	"kratos-template/app/auth/internal/conf"
	"kratos-template/app/auth/internal/data"
	"kratos-template/app/auth/internal/server"
	"kratos-template/app/auth/internal/service"
	"kratos-template/pkg/bootstrap"
	"kratos-template/pkg/config"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
	"kratos-template/pkg/registry"
	"kratos-template/pkg/tracing"
)

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/auth.yaml", "config path")
}

func Run() {
	flag.Parse()

	cfg, err := config.New(flagConf, "config/auth/", "")
	if err != nil {
		log.Fatalf("load config error: %v", err)
		return
	}

	bootstrapCfg, err := loadBootstrapConfig(cfg)
	if err != nil {
		log.Fatalf("load bootstrap config error: %v", err)
		return
	}

	logger, shutdown, err := log.InitFromSettings(provideLogSettings(bootstrapCfg))
	if err != nil {
		log.Fatalf("init log error: %v", err)
		return
	}

	app := fx.New(
		fx.WithLogger(func() fxevent.Logger { return adapter.NewFxAdapter() }),
		fx.Supply(
			fx.Annotate(cfg, fx.As(new(kratosconfig.Config))),
			bootstrapCfg,
			logger,
		),
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
		fx.Provide(func(b *conf.Bootstrap) config.Accessor { return b }),

		fx.Provide(provideRegistrySettings),
		fx.Provide(registry.New),
		fx.Provide(provideTracingSettings),
		fx.Provide(tracing.New),

		fx.Provide(provideServiceInfo),
		fx.Provide(bootstrap.NewKratosApp),

		data.Module,
		biz.Module,
		service.Module,
		server.Module,

		fx.Invoke(func(*kratos.App) {}),
	)

	app.Run()
}

func loadBootstrapConfig(cfg kratosconfig.Config) (*conf.Bootstrap, error) {
	var bc conf.Bootstrap
	if err := cfg.Scan(&bc); err != nil {
		return nil, err
	}
	return &bc, nil
}

func provideLogSettings(cfg *conf.Bootstrap) log.Settings {
	return log.Settings{
		Level:  bootstrap.LogLevelFromConfig(cfg),
		Format: "json",
		Caller: true,
	}
}

type registrySettingsResult struct {
	fx.Out
	Address string `name:"consul_address"`
	Scheme  string `name:"consul_scheme"`
}

func provideRegistrySettings(cfg *conf.Bootstrap) registrySettingsResult {
	settings := bootstrap.RegistryFromConfig(cfg)
	return registrySettingsResult{
		Address: settings.Address,
		Scheme:  settings.Scheme,
	}
}

type tracingSettingsResult struct {
	fx.Out
	OTLPEndpoint string  `name:"otlp_endpoint"`
	SampleRate   float64 `name:"trace_sample_rate"`
}

func provideTracingSettings(cfg *conf.Bootstrap) tracingSettingsResult {
	settings := bootstrap.TracingFromConfig(cfg)
	return tracingSettingsResult{
		OTLPEndpoint: settings.OTLPEndpoint,
		SampleRate:   settings.SampleRate,
	}
}

type serviceInfoResult struct {
	fx.Out
	ServiceID       string            `name:"service_id"`
	ServiceName     string            `name:"service_name"`
	ServiceVersion  string            `name:"service_version"`
	ServiceMetadata map[string]string `name:"service_metadata"`
}

func provideServiceInfo(cfg *conf.Bootstrap) serviceInfoResult {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	info := bootstrap.ServiceInfoFromConfig(cfg, hostname)
	return serviceInfoResult{
		ServiceID:       info.ID,
		ServiceName:     info.Name,
		ServiceVersion:  info.Version,
		ServiceMetadata: info.Metadata,
	}
}

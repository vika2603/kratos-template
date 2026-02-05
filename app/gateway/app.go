package gateway

import (
	"context"
	"flag"
	"os"

	kratosconfig "github.com/go-kratos/kratos/v2/config"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	"kratos-template/app/gateway/biz/module"
	"kratos-template/app/gateway/internal/client"
	"kratos-template/app/gateway/internal/conf"
	"kratos-template/app/gateway/internal/server"
	"kratos-template/app/gateway/pkg/handler"
	pkgauth "kratos-template/pkg/auth"
	"kratos-template/pkg/bootstrap"
	"kratos-template/pkg/config"
	"kratos-template/pkg/log"
	"kratos-template/pkg/registry"
	"kratos-template/pkg/tracing"
)

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/gateway.yaml", "config path")
}

func Run() {
	flag.Parse()

	cfg, err := config.New(flagConf, "config/gateway/", "")
	if err != nil {
		log.Fatalf("load config error: %v", err)
		return
	}

	bootstrapCfg, err := loadBootstrapConfig(cfg)
	if err != nil {
		log.Fatalf("load bootstrap config error: %v", err)
		return
	}

	logger, shutdown, err := log.InitFromSettings(getLogSettings(bootstrapCfg))
	if err != nil {
		log.Fatalf("init log error: %v", err)
		return
	}

	app := fx.New(
		// fx.WithLogger(adapter.NewFxAdapter),
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

		fx.Provide(getRegistrySettings),
		fx.Provide(getTracingSettings),
		fx.Provide(provideServiceInfo),

		fx.Provide(registry.New),
		fx.Provide(tracing.New),

		client.ProvideClients(),

		fx.Provide(server.NewHertzServer, handler.Default),

		fx.Provide(provideAuthConfig),
		pkgauth.Module,

		module.Modules(),

		fx.Invoke(func(lc fx.Lifecycle, s *server.HertzServer, _ trace.TracerProvider) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error { return s.Start(ctx) },
				OnStop:  func(ctx context.Context) error { return s.Stop(ctx) },
			})
		}),
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

func getLogSettings(cfg *conf.Bootstrap) log.Settings {
	level := "info"
	if cfg.Log != nil && cfg.Log.Level != "" {
		level = cfg.Log.Level
	}
	return log.Settings{
		Level:  level,
		Format: "json",
		Caller: true,
	}
}

type registrySettingsResult struct {
	fx.Out
	Address string `name:"consul_address"`
	Scheme  string `name:"consul_scheme"`
}

func getRegistrySettings(cfg *conf.Bootstrap) registrySettingsResult {
	return registrySettingsResult{
		Address: cfg.Registry.Consul.Address,
		Scheme:  cfg.Registry.Consul.Scheme,
	}
}

type tracingSettingsResult struct {
	fx.Out
	OTLPEndpoint string  `name:"otlp_endpoint"`
	SampleRate   float64 `name:"trace_sample_rate"`
}

func getTracingSettings(cfg *conf.Bootstrap) tracingSettingsResult {
	endpoint := ""
	sampleRate := 1.0
	if cfg.Tracing != nil && cfg.Tracing.Otlp != nil {
		endpoint = cfg.Tracing.Otlp.Endpoint
		sampleRate = float64(cfg.Tracing.SampleRate)
	}
	return tracingSettingsResult{
		OTLPEndpoint: endpoint,
		SampleRate:   sampleRate,
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

func provideAuthConfig(cfg *conf.Bootstrap) pkgauth.Config {
	return pkgauth.Config{
		Secret:      cfg.Auth.JwtSecret,
		ExpiryHours: 24,
	}
}

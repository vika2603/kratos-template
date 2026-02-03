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
	"kratos-template/app/gateway/pkg/auth"
	"kratos-template/app/gateway/pkg/handler"
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

	app := fx.New(
		log.FxLogger(),
		config.ProvideWithConsul(flagConf, "config/gateway/", ""),
		fx.Provide(loadBootstrapConfig),
		fx.Provide(func(b *conf.Bootstrap) config.Accessor { return b }),

		fx.Provide(getLoggerSettings),
		fx.Provide(getRegistrySettings),
		fx.Provide(getTracingSettings),
		fx.Provide(provideServiceInfo),

		fx.Provide(log.New),
		fx.Provide(registry.New),
		fx.Provide(tracing.New),

		client.ProvideClients(),

		fx.Provide(server.NewHertzServer, handler.Default),

		fx.Invoke(initAuth),

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

type configResult struct {
	fx.Out
	Bootstrap *conf.Bootstrap
}

func loadBootstrapConfig(cfg kratosconfig.Config) (configResult, error) {
	var bc conf.Bootstrap
	if err := cfg.Scan(&bc); err != nil {
		return configResult{}, err
	}
	return configResult{Bootstrap: &bc}, nil
}

type loggerSettingsResult struct {
	fx.Out
	Level string `name:"log_level"`
	Env   string `name:"env"`
}

func getLoggerSettings(cfg *conf.Bootstrap) loggerSettingsResult {
	return loggerSettingsResult{
		Level: cfg.Log.Level,
		Env:   cfg.Log.Env,
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
	JaegerEndpoint string  `name:"jaeger_endpoint"`
	SampleRate     float64 `name:"trace_sample_rate"`
}

func getTracingSettings(cfg *conf.Bootstrap) tracingSettingsResult {
	endpoint := ""
	sampleRate := 1.0
	if cfg.Tracing != nil && cfg.Tracing.Jaeger != nil {
		endpoint = cfg.Tracing.Jaeger.Endpoint
		sampleRate = float64(cfg.Tracing.SampleRate)
	}
	return tracingSettingsResult{
		JaegerEndpoint: endpoint,
		SampleRate:     sampleRate,
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

func initAuth(cfg *conf.Bootstrap) {
	auth.Init(cfg.Auth.JwtSecret)
}

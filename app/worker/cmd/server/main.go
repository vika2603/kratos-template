package main

import (
	"flag"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"

	"kratos-template/app/worker/internal/activities"
	"kratos-template/app/worker/internal/conf"
	"kratos-template/app/worker/internal/server"
	"kratos-template/app/worker/internal/worker"
	"kratos-template/pkg/bootstrap"
)

var (
	flagConf string
)

func init() {
	flag.StringVar(&flagConf, "conf", "configs/config.yaml", "config path, eg: -conf config.yaml")
}

type ConfigParams struct {
	fx.In
	Config config.Config
}

type ConfigResult struct {
	fx.Out
	Bootstrap *conf.Bootstrap
}

func loadBootstrapConfig(params ConfigParams) (ConfigResult, error) {
	var bc conf.Bootstrap
	if err := params.Config.Scan(&bc); err != nil {
		return ConfigResult{}, err
	}
	return ConfigResult{Bootstrap: &bc}, nil
}

func provideLoggerFromConfig(cfg *conf.Bootstrap) bootstrap.LoggerParams {
	level := "info"
	env := "development"
	if cfg.Log != nil {
		if cfg.Log.Level != "" {
			level = cfg.Log.Level
		}
		if cfg.Log.Env != "" {
			env = cfg.Log.Env
		}
	}
	return bootstrap.LoggerParams{Level: level, Env: env}
}

func provideRegistryFromConfig(cfg *conf.Bootstrap) bootstrap.RegistryParams {
	address := "localhost:8500"
	scheme := "http"
	if cfg.Registry != nil && cfg.Registry.Consul != nil {
		if cfg.Registry.Consul.Address != "" {
			address = cfg.Registry.Consul.Address
		}
		if cfg.Registry.Consul.Scheme != "" {
			scheme = cfg.Registry.Consul.Scheme
		}
	}
	return bootstrap.RegistryParams{Address: address, Scheme: scheme}
}

func provideTracingFromConfig(cfg *conf.Bootstrap) bootstrap.TracingParams {
	endpoint := "http://localhost:14268/api/traces"
	sampleRate := 1.0
	serviceName := "worker"
	serviceVersion := "v1.0.0"
	if cfg.Tracing != nil {
		if cfg.Tracing.Jaeger != nil && cfg.Tracing.Jaeger.Endpoint != "" {
			endpoint = cfg.Tracing.Jaeger.Endpoint
		}
		if cfg.Tracing.SampleRate > 0 {
			sampleRate = cfg.Tracing.SampleRate
		}
	}
	if cfg.Service != nil {
		if cfg.Service.Name != "" {
			serviceName = cfg.Service.Name
		}
		if cfg.Service.Version != "" {
			serviceVersion = cfg.Service.Version
		}
	}
	return bootstrap.TracingParams{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		JaegerEndpoint: endpoint,
		SampleRate:     sampleRate,
	}
}

func main() {
	flag.Parse()

	app := fx.New(
		bootstrap.ProvideConfigWithConsul(flagConf, "config/worker/", ""),
		fx.Provide(loadBootstrapConfig),

		fx.Provide(provideLoggerFromConfig),

		fx.Provide(provideRegistryFromConfig),

		fx.Provide(provideTracingFromConfig),

		bootstrap.ProvideServiceInfo(
			"worker-"+generateID(),
			"worker",
			"v1.0.0",
			nil,
		),

		fx.Provide(activities.NewActivities),
		fx.Provide(worker.NewTemporalWorker),

		server.Module,

		fx.Provide(bootstrap.NewKratosApp),

		fx.Invoke(func(*kratos.App) {}),
	)

	app.Run()
}

func generateID() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
}

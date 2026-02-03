package main

import (
	"flag"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"

	"kratos-template/app/auth/internal/biz"
	"kratos-template/app/auth/internal/conf"
	"kratos-template/app/auth/internal/data"
	"kratos-template/app/auth/internal/server"
	"kratos-template/app/auth/internal/service"
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

func getKratosModule(cfg *conf.Bootstrap) fx.Option {
	loggerCfg := bootstrap.ProvideLoggerFromConfig(cfg)
	registryCfg := bootstrap.ProvideRegistryFromConfig(cfg)
	tracingCfg := bootstrap.ProvideTracingFromConfig(cfg)
	return bootstrap.KratosModule(
		bootstrap.LoggerSettings{Level: loggerCfg.Level, Env: loggerCfg.Env},
		bootstrap.RegistrySettings{Address: registryCfg.Address, Scheme: registryCfg.Scheme},
		bootstrap.TracingSettings{
			ServiceName:    tracingCfg.ServiceName,
			ServiceVersion: tracingCfg.ServiceVersion,
			JaegerEndpoint: tracingCfg.JaegerEndpoint,
			SampleRate:     tracingCfg.SampleRate,
		},
	)
}

func main() {
	flag.Parse()

	app := fx.New(
		bootstrap.ProvideConfigWithConsul(flagConf, "config/auth/", ""),
		fx.Provide(loadBootstrapConfig),

		fx.Provide(func(b *conf.Bootstrap) bootstrap.ConfigAccessor { return b }),
		fx.Provide(getKratosModule),

		fx.Provide(provideServiceInfoFromConfig),

		data.Module,
		biz.Module,
		service.Module,
		server.Module,

		fx.Provide(bootstrap.NewKratosApp),

		fx.Invoke(func(*kratos.App) {}),
	)

	app.Run()
}

type serviceInfoResult struct {
	fx.Out
	ServiceID       string            `name:"service_id"`
	ServiceName     string            `name:"service_name"`
	ServiceVersion  string            `name:"service_version"`
	ServiceMetadata map[string]string `name:"service_metadata"`
}

func provideServiceInfoFromConfig(cfg *conf.Bootstrap) serviceInfoResult {
	name := "auth"
	version := "v1.0.0"
	if cfg.Service != nil {
		if cfg.Service.Name != "" {
			name = cfg.Service.Name
		}
		if cfg.Service.Version != "" {
			version = cfg.Service.Version
		}
	}
	return serviceInfoResult{
		ServiceID:       name + "-" + generateID(),
		ServiceName:     name,
		ServiceVersion:  version,
		ServiceMetadata: nil,
	}
}

func generateID() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
}

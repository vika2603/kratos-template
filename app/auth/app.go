package auth

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

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/auth.yaml", "config path")
}

func Run() {
	flag.Parse()

	app := fx.New(
		bootstrap.FxLogger(),
		bootstrap.ProvideConfigWithConsul(flagConf, "config/auth/", ""),
		fx.Provide(loadBootstrapConfig),

		fx.Provide(func(b *conf.Bootstrap) bootstrap.ConfigAccessor { return b }),

		fx.Provide(getKratosModule),

		fx.Provide(provideServiceInfo),

		data.Module,
		biz.Module,
		service.Module,
		server.Module,

		fx.Invoke(func(*kratos.App) {}),
	)

	app.Run()
}

type configResult struct {
	fx.Out
	Bootstrap *conf.Bootstrap
}

func loadBootstrapConfig(cfg config.Config) (configResult, error) {
	var bc conf.Bootstrap
	if err := cfg.Scan(&bc); err != nil {
		return configResult{}, err
	}
	return configResult{Bootstrap: &bc}, nil
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
	params := bootstrap.ProvideServiceInfoFromConfig(cfg, hostname)
	return serviceInfoResult{
		ServiceID:       params.ServiceID,
		ServiceName:     params.ServiceName,
		ServiceVersion:  params.ServiceVersion,
		ServiceMetadata: params.ServiceMetadata,
	}
}

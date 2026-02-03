package user

import (
	"flag"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"

	"kratos-template/app/user/internal/biz"
	"kratos-template/app/user/internal/conf"
	"kratos-template/app/user/internal/data"
	"kratos-template/app/user/internal/server"
	"kratos-template/app/user/internal/service"
	"kratos-template/pkg/bootstrap"
)

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/user.yaml", "config path")
}

func Run() {
	flag.Parse()

	app := fx.New(
		bootstrap.FxLogger(),
		bootstrap.ProvideConfigWithConsul(flagConf, "config/user/", ""),
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

type serviceInfoResult struct {
	fx.Out
	ServiceID       string            `name:"service_id"`
	ServiceName     string            `name:"service_name"`
	ServiceVersion  string            `name:"service_version"`
	ServiceMetadata map[string]string `name:"service_metadata"`
}

func provideServiceInfo(cfg *conf.Bootstrap) serviceInfoResult {
	params := bootstrap.ProvideServiceInfoFromConfig(cfg, getHostname())
	return serviceInfoResult{
		ServiceID:       params.ServiceID,
		ServiceName:     params.ServiceName,
		ServiceVersion:  params.ServiceVersion,
		ServiceMetadata: params.ServiceMetadata,
	}
}

func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
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

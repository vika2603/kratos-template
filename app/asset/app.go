package asset

import (
	"flag"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"

	"kratos-template/app/asset/internal/biz"
	"kratos-template/app/asset/internal/conf"
	"kratos-template/app/asset/internal/data"
	"kratos-template/app/asset/internal/server"
	"kratos-template/app/asset/internal/service"
	"kratos-template/pkg/bootstrap"
)

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/asset.yaml", "config path")
}

func Run() {
	flag.Parse()

	app := fx.New(
		bootstrap.FxLogger(),
		bootstrap.ProvideConfigWithConsul(flagConf, "config/asset/", ""),
		fx.Provide(func(cfg config.Config) (*conf.Bootstrap, error) {
			var bc conf.Bootstrap
			if err := cfg.Scan(&bc); err != nil {
				return nil, err
			}
			return &bc, nil
		}),
		bootstrap.ProvideConfigAccessor[*conf.Bootstrap](),
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

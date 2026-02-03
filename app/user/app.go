package user

import (
	"flag"
	"os"

	"github.com/go-kratos/kratos/v2"
	kratosconfig "github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"

	"kratos-template/app/user/internal/biz"
	"kratos-template/app/user/internal/conf"
	"kratos-template/app/user/internal/data"
	"kratos-template/app/user/internal/server"
	"kratos-template/app/user/internal/service"
	"kratos-template/pkg/bootstrap"
	"kratos-template/pkg/config"
	"kratos-template/pkg/log"
	"kratos-template/pkg/registry"
	"kratos-template/pkg/tracing"
)

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/user.yaml", "config path")
}

func Run() {
	flag.Parse()

	app := fx.New(
		log.FxLogger(),
		config.ProvideWithConsul(flagConf, "config/user/", ""),
		fx.Provide(loadBootstrapConfig),

		fx.Provide(func(b *conf.Bootstrap) config.Accessor { return b }),

		fx.Provide(getModule),

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

func loadBootstrapConfig(cfg kratosconfig.Config) (configResult, error) {
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
	info := bootstrap.ServiceInfoFromConfig(cfg, getHostname())
	return serviceInfoResult{
		ServiceID:       info.ID,
		ServiceName:     info.Name,
		ServiceVersion:  info.Version,
		ServiceMetadata: info.Metadata,
	}
}

func getHostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
}

func getModule(cfg *conf.Bootstrap) fx.Option {
	loggerCfg := bootstrap.LoggerFromConfig(cfg)
	registryCfg := bootstrap.RegistryFromConfig(cfg)
	tracingCfg := bootstrap.TracingFromConfig(cfg)
	return bootstrap.Module(
		log.Settings{Level: loggerCfg.Level, Env: loggerCfg.Env},
		registry.Settings{Address: registryCfg.Address, Scheme: registryCfg.Scheme},
		tracing.Settings{
			ServiceName:    tracingCfg.ServiceName,
			ServiceVersion: tracingCfg.ServiceVersion,
			JaegerEndpoint: tracingCfg.JaegerEndpoint,
			SampleRate:     tracingCfg.SampleRate,
		},
	)
}

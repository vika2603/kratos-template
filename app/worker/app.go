package worker

import (
	"flag"
	"os"

	"github.com/go-kratos/kratos/v2"
	kratosconfig "github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"

	"kratos-template/app/worker/internal/activities"
	"kratos-template/app/worker/internal/conf"
	"kratos-template/app/worker/internal/server"
	"kratos-template/app/worker/internal/worker"
	"kratos-template/pkg/bootstrap"
	"kratos-template/pkg/config"
	"kratos-template/pkg/log"
	"kratos-template/pkg/registry"
	"kratos-template/pkg/tracing"
)

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/worker.yaml", "config path")
}

func Run() {
	flag.Parse()

	app := fx.New(
		log.FxLogger(),
		config.ProvideWithConsul(flagConf, "config/worker/", ""),
		fx.Provide(loadBootstrapConfig),
		config.ProvideAccessor[*conf.Bootstrap](),
		fx.Provide(getModule),
		fx.Provide(provideServiceInfo),
		fx.Provide(activities.NewActivities),
		worker.Module,
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

func getModule(cfg config.Accessor) fx.Option {
	l := bootstrap.LoggerFromConfig(cfg)
	r := bootstrap.RegistryFromConfig(cfg)
	t := bootstrap.TracingFromConfig(cfg)
	return bootstrap.Module(
		log.Settings{Level: l.Level, Env: l.Env},
		registry.Settings{Address: r.Address, Scheme: r.Scheme},
		tracing.Settings{
			ServiceName:    t.ServiceName,
			ServiceVersion: t.ServiceVersion,
			JaegerEndpoint: t.JaegerEndpoint,
			SampleRate:     t.SampleRate,
		},
	)
}

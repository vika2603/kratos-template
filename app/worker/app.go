package worker
import (
	"flag"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"
	"kratos-template/app/worker/internal/activities"
	"kratos-template/app/worker/internal/conf"
	"kratos-template/app/worker/internal/server"
	"kratos-template/app/worker/internal/worker"
	"kratos-template/pkg/bootstrap"
	"os"
)
var flagConf string
func init() {
	flag.StringVar(&flagConf, "conf", "configs/worker.yaml", "config path")
}
func Run() {
	flag.Parse()
	app := fx.New(
		bootstrap.ProvideConfigWithConsul(flagConf, "config/worker/", ""),
		fx.Provide(loadBootstrapConfig),
		bootstrap.ProvideConfigAccessor[*conf.Bootstrap](),
		fx.Provide(getKratosModule),
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
func getKratosModule(cfg bootstrap.ConfigAccessor) fx.Option {
	l := bootstrap.ProvideLoggerFromConfig(cfg)
	r := bootstrap.ProvideRegistryFromConfig(cfg)
	t := bootstrap.ProvideTracingFromConfig(cfg)
	return bootstrap.KratosModule(
		bootstrap.LoggerSettings{Level: l.Level, Env: l.Env},
		bootstrap.RegistrySettings{Address: r.Address, Scheme: r.Scheme},
		bootstrap.TracingSettings{
			ServiceName:    t.ServiceName,
			ServiceVersion: t.ServiceVersion,
			JaegerEndpoint: t.JaegerEndpoint,
			SampleRate:     t.SampleRate,
		},
	)
}

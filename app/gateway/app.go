package gateway

import (
	"context"
	"flag"

	kratosconfig "github.com/go-kratos/kratos/v2/config"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"kratos-template/app/gateway/biz/module"
	"kratos-template/app/gateway/internal/client"
	"kratos-template/app/gateway/internal/conf"
	"kratos-template/app/gateway/internal/server"
	"kratos-template/app/gateway/pkg/handler"
	authmw "kratos-template/app/gateway/pkg/middleware/auth"
	pkgauth "kratos-template/pkg/auth"
	"kratos-template/pkg/bootstrap"
	"kratos-template/pkg/config"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
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

	bootstrapCfg, err := bootstrap.LoadConfig[conf.Bootstrap](cfg)
	if err != nil {
		log.Fatalf("load bootstrap config error: %v", err)
		return
	}

	logger, shutdown, err := log.InitFromSettings(bootstrap.ProvideLogSettings(bootstrapCfg))
	if err != nil {
		log.Fatalf("init log error: %v", err)
		return
	}

	app := fx.New(
		fx.WithLogger(func() fxevent.Logger { return adapter.NewFxAdapter() }),
		fx.Supply(
			fx.Annotate(cfg, fx.As(new(kratosconfig.Config))),
			bootstrapCfg,
			logger,
		),
		bootstrap.CommonLifecycleOptions(shutdown),
		config.ProvideAccessor[*conf.Bootstrap](),

		fx.Provide(bootstrap.ProvideRegistrySettings),
		fx.Provide(registry.New),
		fx.Provide(bootstrap.ProvideTracingSettings),
		fx.Provide(tracing.New),
		fx.Provide(bootstrap.ProvideServiceInfo),

		client.ProvideClients(),

		fx.Provide(server.NewHertzServer, handler.Default),

		fx.Provide(provideAuthConfig),
		pkgauth.Module,

		fx.Invoke(func(mgr *pkgauth.JWTManager) {
			authmw.Init(mgr)
		}),

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

func provideAuthConfig(cfg *conf.Bootstrap) pkgauth.Config {
	return pkgauth.Config{
		Secret:      cfg.Auth.JwtSecret,
		ExpiryHours: 24,
	}
}

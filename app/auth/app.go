package auth

import (
	"flag"

	"github.com/go-kratos/kratos/v2"
	kratosconfig "github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"kratos-template/app/auth/internal/biz"
	"kratos-template/app/auth/internal/conf"
	"kratos-template/app/auth/internal/data"
	"kratos-template/app/auth/internal/server"
	"kratos-template/app/auth/internal/service"
	"kratos-template/pkg/bootstrap"
	"kratos-template/pkg/config"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
	"kratos-template/pkg/registry"
	"kratos-template/pkg/tracing"
)

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/auth.yaml", "config path")
}

func Run() {
	flag.Parse()

	cfg, err := config.New(flagConf, "config/auth/", "")
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
		fx.Provide(bootstrap.NewKratosApp),

		data.Module,
		biz.Module,
		service.Module,
		server.Module,

		fx.Invoke(func(*kratos.App) {}),
	)

	app.Run()
}

package bootstrap

import (
	"context"

	kratosconfig "github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
)

func Run[T any](configPath, consulPath string, opts ...fx.Option) {
	cfg, err := NewConfig(configPath, consulPath, "")
	if err != nil {
		log.Fatalf("load config error: %v", err)
	}

	bc, err := LoadConfig[T](cfg)
	if err != nil {
		log.Fatalf("load bootstrap config error: %v", err)
	}

	cc, err := ScanCommonConfig(cfg)
	if err != nil {
		log.Fatalf("scan common config error: %v", err)
	}

	logger, shutdown, err := log.Init(ProvideLogSettings(cc))
	if err != nil {
		log.Fatalf("init log error: %v", err)
	}

	tracerShutdown, err := InitTracer(cc.GetService().GetName(), cc.GetService().GetVersion())
	if err != nil {
		log.Errorf("init tracer error: %v", err)
		tracerShutdown = func(context.Context) error { return nil }
	}

	allOpts := []fx.Option{
		fx.WithLogger(func() fxevent.Logger { return adapter.NewFxAdapter() }),
		fx.Supply(
			fx.Annotate(cfg, fx.As(new(kratosconfig.Config))),
			bc,
			cc,
			logger,
		),
		CommonLifecycleOptions(shutdown),
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{OnStop: tracerShutdown})
		}),
		CommonProviders(),
	}
	allOpts = append(allOpts, opts...)

	fx.New(allOpts...).Run()
}

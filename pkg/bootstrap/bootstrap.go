package bootstrap

import (
	kratosconfig "github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"

	"kratos-template/pkg/log"
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

	allOpts := []fx.Option{
		// fx.WithLogger(func() fxevent.Logger { return adapter.NewFxAdapter() }),
		fx.Supply(
			fx.Annotate(cfg, fx.As(new(kratosconfig.Config))),
			bc,
			cc,
			logger,
		),
		CommonLifecycleOptions(shutdown),
		CommonProviders(),
	}
	allOpts = append(allOpts, opts...)

	fx.New(allOpts...).Run()
}

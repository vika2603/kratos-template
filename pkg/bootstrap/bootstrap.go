package bootstrap

import (
	"context"
	"flag"
	"fmt"
	"kratos-template/pkg/conf"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"

	kratosconfig "github.com/go-kratos/kratos/v2/config"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

// Version is stamped at build time via -ldflags (see Makefile); "dev" locally.
var Version = "dev"

// Run owns flag parsing — add shared flags here, not per-service.
func Run[T any](name string, opts ...fx.Option) {
	configPath := flag.String("conf", "configs/"+name+".yaml", "config file path")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(name, Version)
		return
	}

	cfg, err := NewConfig(*configPath, "config/"+name+"/")
	if err != nil {
		log.Fatalf("load config error: %v", err)
	}

	bc, err := LoadConfig[T](cfg)
	if err != nil {
		log.Fatalf("load bootstrap config error: %v", err)
	}

	sc, err := LoadConfig[conf.Shared](cfg)
	if err != nil {
		log.Fatalf("load shared config error: %v", err)
	}

	logger, shutdown, err := log.Init(ProvideLogSettings(sc))
	if err != nil {
		log.Fatalf("init log error: %v", err)
	}

	tracerShutdown, err := InitTracer(name, Version)
	if err != nil {
		log.Errorf("init tracer error: %v", err)
		tracerShutdown = func(context.Context) error { return nil }
	}

	registry, metricsShutdown, err := InitMetrics(name, Version)
	if err != nil {
		log.Errorf("init metrics error: %v", err)
		registry = prometheus.NewRegistry()
		metricsShutdown = func(context.Context) error { return nil }
	}

	allOpts := []fx.Option{
		fx.WithLogger(adapter.NewFxAdapter),
		fx.Supply(
			fx.Annotate(cfg, fx.As(new(kratosconfig.Config))),
			bc,
			sc,
			logger,
			registry,
		),
		SharedLifecycleOptions(shutdown),
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{OnStop: tracerShutdown})
			lc.Append(fx.Hook{OnStop: metricsShutdown})
		}),
		SharedProviders(name, Version),
	}
	allOpts = append(allOpts, opts...)

	fx.New(allOpts...).Run()
}

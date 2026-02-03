package bootstrap

import (
	"go.uber.org/fx"

	"kratos-template/pkg/log"
	"kratos-template/pkg/registry"
	"kratos-template/pkg/tracing"
)

func Module(logger log.Settings, reg registry.Settings, trace tracing.Settings) fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func() string { return logger.Level },
				fx.ResultTags(`name:"log_level"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() string { return logger.Env },
				fx.ResultTags(`name:"env"`),
			),
		),
		fx.Provide(log.New),

		fx.Provide(
			fx.Annotate(
				func() string { return reg.Address },
				fx.ResultTags(`name:"consul_address"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() string { return reg.Scheme },
				fx.ResultTags(`name:"consul_scheme"`),
			),
		),
		fx.Provide(registry.New),

		fx.Provide(
			fx.Annotate(
				func() string { return trace.JaegerEndpoint },
				fx.ResultTags(`name:"jaeger_endpoint"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() float64 { return trace.SampleRate },
				fx.ResultTags(`name:"trace_sample_rate"`),
			),
		),
		fx.Provide(tracing.New),

		fx.Provide(NewKratosApp),
	)
}

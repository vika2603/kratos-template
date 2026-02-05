package bootstrap

import (
	"go.uber.org/fx"

	"kratos-template/pkg/log"
	"kratos-template/pkg/registry"
	"kratos-template/pkg/tracing"
)

func Module(logLevel string, reg registry.Settings, trace tracing.Settings) fx.Option {
	return fx.Options(
		log.ProvideWithSettings(log.Settings{
			Level:  logLevel,
			Format: "json",
			Caller: true,
		}),

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
				func() string { return trace.OTLPEndpoint },
				fx.ResultTags(`name:"otlp_endpoint"`),
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

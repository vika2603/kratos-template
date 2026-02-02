package bootstrap

import "go.uber.org/fx"

// KratosModule consolidates Logger, Registry, and Tracing providers into a single reusable fx.Option.
// It does NOT provide ServiceInfo (service_id, service_name, service_version, service_metadata) to avoid
// duplicate named dependency conflicts with service-specific provideServiceInfo().
func KratosModule(logger LoggerSettings, registry RegistrySettings, tracing TracingSettings) fx.Option {
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
		fx.Provide(NewLogger),

		fx.Provide(
			fx.Annotate(
				func() string { return registry.Address },
				fx.ResultTags(`name:"consul_address"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() string { return registry.Scheme },
				fx.ResultTags(`name:"consul_scheme"`),
			),
		),
		fx.Provide(NewRegistry),

		fx.Provide(
			fx.Annotate(
				func() string { return tracing.JaegerEndpoint },
				fx.ResultTags(`name:"jaeger_endpoint"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() float64 { return tracing.SampleRate },
				fx.ResultTags(`name:"trace_sample_rate"`),
			),
		),
		fx.Provide(NewTracing),

		fx.Provide(NewKratosApp),
	)
}

package bootstrap

import "go.uber.org/fx"

// LoggerSettings holds logger configuration values
type LoggerSettings struct {
	Level string
	Env   string
}

// RegistrySettings holds service registry (Consul) configuration values
type RegistrySettings struct {
	Address string
	Scheme  string
}

// TracingSettings holds distributed tracing (Jaeger) configuration values
type TracingSettings struct {
	ServiceName    string
	ServiceVersion string
	JaegerEndpoint string
	SampleRate     float64
}

// ServiceInfoSettings holds service identification information
type ServiceInfoSettings struct {
	ID       string
	Name     string
	Version  string
	Metadata map[string]string
}

// ProvideConfigAccessor returns an fx.Option that provides a ConfigAccessor
// from any type that implements the ConfigAccessor interface.
// This is a helper to reduce boilerplate in service modules.
func ProvideConfigAccessor[T ConfigAccessor]() fx.Option {
	return fx.Provide(func(cfg T) ConfigAccessor {
		return cfg
	})
}

package config

import "go.uber.org/fx"

type Accessor interface {
	GetServiceName() string
	GetServiceVersion() string
	GetConsulAddress() string
	GetConsulScheme() string
	GetOTLPEndpoint() string
	GetTracingSampleRate() float64
	GetLogLevel() string
	GetLogEnv() string
}

func ProvideAccessor[T Accessor]() fx.Option {
	return fx.Provide(func(cfg T) Accessor {
		return cfg
	})
}

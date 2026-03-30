package bootstrap

import (
	"context"

	kratosconfig "github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"

	"kratos-template/pkg/conf"
	"kratos-template/pkg/log"
	"kratos-template/pkg/registry"
)

type RegistrySettingsResult struct {
	fx.Out
	Address string `name:"consul_address"`
	Scheme  string `name:"consul_scheme"`
}

type ServiceInfoResult struct {
	fx.Out
	ServiceID       string            `name:"service_id"`
	ServiceName     string            `name:"service_name"`
	ServiceVersion  string            `name:"service_version"`
	ServiceMetadata map[string]string `name:"service_metadata"`
}

func ProvideLogSettings(cfg *conf.CommonConfig) log.Config {
	level := cfg.GetLog().GetLevel()
	if level == "" {
		level = "info"
	}

	env := cfg.GetLog().GetEnv()
	format := "json"
	development := false
	if env == "development" {
		format = "console"
		development = true
	}

	return log.Config{
		Level:       level,
		Format:      format,
		Development: development,
		Caller:      true,
	}
}

func ProvideRegistrySettings(cfg *conf.CommonConfig) RegistrySettingsResult {
	address := cfg.GetRegistry().GetConsul().GetAddress()
	if address == "" {
		address = "localhost:8500"
	}
	scheme := cfg.GetRegistry().GetConsul().GetScheme()
	if scheme == "" {
		scheme = "http"
	}
	return RegistrySettingsResult{
		Address: address,
		Scheme:  scheme,
	}
}

func ProvideServiceInfo(cfg *conf.CommonConfig) ServiceInfoResult {
	name := cfg.GetService().GetName()
	if name == "" {
		name = "unknown"
	}
	version := cfg.GetService().GetVersion()
	if version == "" {
		version = "v0.0.0"
	}
	hostname := Hostname()
	return ServiceInfoResult{
		ServiceID:       name + "-" + hostname,
		ServiceName:     name,
		ServiceVersion:  version,
		ServiceMetadata: nil,
	}
}

func CommonProviders() fx.Option {
	return fx.Options(
		fx.Provide(ProvideRegistrySettings),
		fx.Provide(registry.New),
		fx.Provide(ProvideServiceInfo),
	)
}

func CommonLifecycleOptions(shutdown func(context.Context) error) fx.Option {
	return fx.Options(
		fx.Invoke(func(lc fx.Lifecycle, c kratosconfig.Config) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					return c.Close()
				},
			})
		}),
		fx.Invoke(func(lc fx.Lifecycle) {
			lc.Append(fx.Hook{OnStop: shutdown})
		}),
	)
}

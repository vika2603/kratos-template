package bootstrap

import (
	"go.uber.org/fx"
	"kratos-template/pkg/config"
	"kratos-template/pkg/log"
)

type RegistrySettingsResult struct {
	fx.Out
	Address string `name:"consul_address"`
	Scheme  string `name:"consul_scheme"`
}

type TracingSettingsResult struct {
	fx.Out
	OTLPEndpoint string  `name:"otlp_endpoint"`
	SampleRate   float64 `name:"trace_sample_rate"`
}

type ServiceInfoResult struct {
	fx.Out
	ServiceID       string            `name:"service_id"`
	ServiceName     string            `name:"service_name"`
	ServiceVersion  string            `name:"service_version"`
	ServiceMetadata map[string]string `name:"service_metadata"`
}

func ProvideLogSettings(cfg config.Accessor) log.Settings {
	return log.Settings{
		Level:  LogLevelFromConfig(cfg),
		Format: "json",
		Caller: true,
	}
}

func ProvideRegistrySettings(cfg config.Accessor) RegistrySettingsResult {
	settings := RegistryFromConfig(cfg)
	return RegistrySettingsResult{
		Address: settings.Address,
		Scheme:  settings.Scheme,
	}
}

func ProvideTracingSettings(cfg config.Accessor) TracingSettingsResult {
	settings := TracingFromConfig(cfg)
	return TracingSettingsResult{
		OTLPEndpoint: settings.OTLPEndpoint,
		SampleRate:   settings.SampleRate,
	}
}

func ProvideServiceInfo(cfg config.Accessor) ServiceInfoResult {
	info := ServiceInfoFromConfig(cfg, Hostname())
	return ServiceInfoResult{
		ServiceID:       info.ID,
		ServiceName:     info.Name,
		ServiceVersion:  info.Version,
		ServiceMetadata: info.Metadata,
	}
}

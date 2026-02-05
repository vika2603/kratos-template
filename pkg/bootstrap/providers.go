package bootstrap

import (
	"kratos-template/pkg/config"
	"kratos-template/pkg/registry"
	"kratos-template/pkg/tracing"
)

func LogLevelFromConfig(cfg config.Accessor) string {
	if l := cfg.GetLogLevel(); l != "" {
		return l
	}
	return "info"
}

func RegistryFromConfig(cfg config.Accessor) registry.Settings {
	address := "localhost:8500"
	scheme := "http"
	if a := cfg.GetConsulAddress(); a != "" {
		address = a
	}
	if s := cfg.GetConsulScheme(); s != "" {
		scheme = s
	}
	return registry.Settings{Address: address, Scheme: scheme}
}

func TracingFromConfig(cfg config.Accessor) tracing.Settings {
	endpoint := "localhost:4317"
	sampleRate := 1.0
	serviceName := cfg.GetServiceName()
	serviceVersion := cfg.GetServiceVersion()

	if serviceName == "" {
		serviceName = "unknown"
	}
	if serviceVersion == "" {
		serviceVersion = "v0.0.0"
	}

	if e := cfg.GetOTLPEndpoint(); e != "" {
		endpoint = e
	}
	if r := cfg.GetTracingSampleRate(); r > 0 {
		sampleRate = r
	}

	return tracing.Settings{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		OTLPEndpoint:   endpoint,
		SampleRate:     sampleRate,
	}
}

type ServiceInfo struct {
	ID       string
	Name     string
	Version  string
	Metadata map[string]string
}

func ServiceInfoFromConfig(cfg config.Accessor, idSuffix string) ServiceInfo {
	name := cfg.GetServiceName()
	version := cfg.GetServiceVersion()

	if name == "" {
		name = "unknown"
	}
	if version == "" {
		version = "v0.0.0"
	}
	if idSuffix == "" {
		idSuffix = "unknown"
	}

	return ServiceInfo{
		ID:       name + "-" + idSuffix,
		Name:     name,
		Version:  version,
		Metadata: nil,
	}
}

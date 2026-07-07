package bootstrap

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// InitMetrics wires the global OTEL meter provider to a private Prometheus
// registry; pkg/metrics.Server exposes it for scraping.
func InitMetrics(serviceName, version string) (*prometheus.Registry, func(context.Context) error, error) {
	registry := prometheus.NewRegistry()

	exporter, err := otelprom.New(otelprom.WithRegisterer(registry))
	if err != nil {
		return nil, nil, err
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(resource.NewSchemaless(
			attribute.String("service.name", serviceName),
			attribute.String("service.version", version),
		)),
	)
	otel.SetMeterProvider(mp)

	return registry, mp.Shutdown, nil
}

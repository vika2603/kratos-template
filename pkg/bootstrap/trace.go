package bootstrap

import (
	"context"
	"os"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

// InitTracer enables OTLP/gRPC tracing when OTEL_EXPORTER_OTLP_ENDPOINT is set;
// otherwise it returns a no-op shutdown. The returned func is called on app stop.
func InitTracer(serviceName, version string) (func(context.Context) error, error) {
	noop := func(context.Context) error { return nil }

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		return noop, nil
	}
	// gRPC exporter wants host:port.
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	exp, err := otlptracegrpc.New(context.Background(),
		tracerExporterOptions(endpoint)...,
	)
	if err != nil {
		return nil, err
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(traceSampleRatio()))),
		tracesdk.WithResource(resource.NewSchemaless(
			attribute.String("service.name", serviceName),
			attribute.String("service.version", version),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

func tracerExporterOptions(endpoint string) []otlptracegrpc.Option {
	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(endpoint)}
	insecure := true
	if raw := os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"); raw != "" {
		if parsed, err := strconv.ParseBool(raw); err == nil {
			insecure = parsed
		}
	}
	if insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	return opts
}

func traceSampleRatio() float64 {
	ratio := 1.0
	if raw := os.Getenv("OTEL_TRACES_SAMPLER_ARG"); raw != "" {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil {
			ratio = parsed
		}
	}
	if ratio < 0 {
		return 0
	}
	if ratio > 1 {
		return 1
	}
	return ratio
}

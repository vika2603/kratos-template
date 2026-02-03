package tracing

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

type Settings struct {
	ServiceName    string
	ServiceVersion string
	JaegerEndpoint string
	SampleRate     float64
}

type Params struct {
	fx.In
	ServiceName    string  `name:"service_name"`
	ServiceVersion string  `name:"service_version"`
	JaegerEndpoint string  `name:"jaeger_endpoint"`
	SampleRate     float64 `name:"trace_sample_rate" optional:"true"`
}

type Result struct {
	fx.Out
	TracerProvider trace.TracerProvider
	Shutdown       func(context.Context) error `name:"tracer_shutdown"`
}

func New(lc fx.Lifecycle, params Params) (Result, error) {
	endpoint := os.Getenv("JAEGER_ENDPOINT")
	if endpoint == "" {
		endpoint = params.JaegerEndpoint
	}
	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)),
	)
	if err != nil {
		return Result{}, err
	}

	sampleRate := 1.0
	if params.SampleRate > 0 && params.SampleRate <= 1.0 {
		sampleRate = params.SampleRate
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(params.ServiceName),
			semconv.ServiceVersion(params.ServiceVersion),
		)),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(sampleRate)),
	)

	otel.SetTracerProvider(tp)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	})

	return Result{
		TracerProvider: tp,
		Shutdown:       tp.Shutdown,
	}, nil
}

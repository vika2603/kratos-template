package tracing

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"

	"kratos-template/pkg/log"
)

type Settings struct {
	ServiceName    string
	ServiceVersion string
	OTLPEndpoint   string
	SampleRate     float64
}

type Params struct {
	fx.In
	ServiceName    string  `name:"service_name"`
	ServiceVersion string  `name:"service_version"`
	OTLPEndpoint   string  `name:"otlp_endpoint"`
	SampleRate     float64 `name:"trace_sample_rate" optional:"true"`
}

type Result struct {
	fx.Out
	TracerProvider trace.TracerProvider
	Shutdown       func(context.Context) error `name:"tracer_shutdown"`
}

func New(lc fx.Lifecycle, params Params) (Result, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = params.OTLPEndpoint
	}

	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Errorf("failed to init otlp exporter: %v", err)
		exporter = nil
	}

	sampleRate := 1.0
	if params.SampleRate > 0 && params.SampleRate <= 1.0 {
		sampleRate = params.SampleRate
	}

	options := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(params.ServiceName),
			semconv.ServiceVersion(params.ServiceVersion),
		)),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(sampleRate)),
	}
	if exporter != nil {
		options = append(options, sdktrace.WithBatcher(exporter))
	}

	tp := sdktrace.NewTracerProvider(options...)

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

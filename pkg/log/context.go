package log

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func WithContextLogger(ctx context.Context, logger *zap.Logger) *zap.Logger {
	if logger == nil {
		return logger
	}
	if fields := extractContextFields(ctx); len(fields) > 0 {
		return logger.With(fields...)
	}
	return logger
}

func extractContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	if tr, ok := transport.FromServerContext(ctx); ok {
		if requestID := tr.RequestHeader().Get("X-Request-ID"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}
		fields = append(fields, zap.String("operation", tr.Operation()))
	}
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		fields = append(fields,
			zap.String("trace_id", spanCtx.TraceID().String()),
			zap.String("span_id", spanCtx.SpanID().String()),
		)
	}

	return fields
}

package log

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport"
	"go.opentelemetry.io/otel/trace"
)

func extractContextFields(ctx context.Context) []Field {
	var fields []Field

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		fields = append(fields, String("trace_id", span.SpanContext().TraceID().String()))
	}
	if span.SpanContext().HasSpanID() {
		fields = append(fields, String("span_id", span.SpanContext().SpanID().String()))
	}

	if tr, ok := transport.FromServerContext(ctx); ok {
		if requestID := tr.RequestHeader().Get("X-Request-ID"); requestID != "" {
			fields = append(fields, String("request_id", requestID))
		}
		fields = append(fields, String("operation", tr.Operation()))
	}

	return fields
}

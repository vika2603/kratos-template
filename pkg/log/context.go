package log

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport"
	"go.uber.org/zap"
)

func WithContextLogger(logger *zap.Logger, ctx context.Context) *zap.Logger {
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

	return fields
}

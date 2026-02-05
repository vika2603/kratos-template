package adapter

import (
	kratoslog "github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"

	"kratos-template/pkg/log"
)

// KratosAdapter adapts our Logger to Kratos log.Logger interface.
type KratosAdapter struct {
	logger *zap.Logger
}

// KratosGlobalAdapter adapts the global logger to Kratos log.Logger interface.
type KratosGlobalAdapter struct{}

// NewKratosAdapter creates a new Kratos adapter.
func NewKratosAdapter(logger *zap.Logger) kratoslog.Logger {
	return &KratosAdapter{logger: logger}
}

// NewKratosGlobalAdapter creates a new adapter bound to the global logger.
func NewKratosGlobalAdapter() kratoslog.Logger {
	return &KratosGlobalAdapter{}
}

// Log implements kratos log.Logger interface.
// Kratos passes key-value pairs: Log(level, "key1", val1, "key2", val2, ...)
func (a *KratosAdapter) Log(level kratoslog.Level, keyvals ...interface{}) error {
	msg, fields := a.parseKeyvals(keyvals)

	switch level {
	case kratoslog.LevelDebug:
		a.logger.Debug(msg, fields...)
	case kratoslog.LevelInfo:
		a.logger.Info(msg, fields...)
	case kratoslog.LevelWarn:
		a.logger.Warn(msg, fields...)
	case kratoslog.LevelError:
		a.logger.Error(msg, fields...)
	case kratoslog.LevelFatal:
		a.logger.Fatal(msg, fields...)
	default:
		a.logger.Info(msg, fields...)
	}

	return nil
}

func (a *KratosGlobalAdapter) Log(level kratoslog.Level, keyvals ...interface{}) error {
	return (&KratosAdapter{logger: log.L()}).Log(level, keyvals...)
}

// parseKeyvals extracts message and fields from Kratos keyvals.
// Kratos convention: keyvals are pairs of (key, value).
// If "msg" key exists, use its value as the message.
func (a *KratosAdapter) parseKeyvals(keyvals []interface{}) (string, []log.Field) {
	if len(keyvals) == 0 {
		return "", nil
	}

	msg := ""
	fields := make([]log.Field, 0, len(keyvals)/2)

	for i := 0; i+1 < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}

		val := keyvals[i+1]

		if key == "msg" || key == "message" {
			if s, ok := val.(string); ok {
				msg = s
				continue
			}
		}

		fields = append(fields, log.Any(key, val))
	}

	return msg, fields
}

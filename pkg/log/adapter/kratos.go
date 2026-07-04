package adapter

import (
	"kratos-template/pkg/log"

	kratoslog "github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
)

type KratosAdapter struct {
	logger *zap.Logger
}

type KratosGlobalAdapter struct{}

func NewKratosAdapter(logger *zap.Logger) kratoslog.Logger {
	return &KratosAdapter{logger: logger}
}

func NewKratosGlobalAdapter() kratoslog.Logger {
	return &KratosGlobalAdapter{}
}

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

func (a *KratosAdapter) parseKeyvals(keyvals []interface{}) (string, []zap.Field) {
	if len(keyvals) == 0 {
		return "", nil
	}

	msg := ""
	fields := make([]zap.Field, 0, len(keyvals)/2)

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

		fields = append(fields, zap.Any(key, val))
	}

	return msg, fields
}

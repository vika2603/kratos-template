package log

import (
	"context"
	"net/http"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level = zapcore.Level

const (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	WarnLevel  = zapcore.WarnLevel
	ErrorLevel = zapcore.ErrorLevel
	FatalLevel = zapcore.FatalLevel
)

var (
	globalLogger       atomic.Pointer[zap.Logger]
	globalCallerLogger atomic.Pointer[zap.Logger]
	globalLevel        atomic.Pointer[zap.AtomicLevel]
	globalFlush        atomic.Value
)

func init() {
	logger, level, flush := NewDefault()
	globalLogger.Store(logger)
	globalCallerLogger.Store(logger.WithOptions(zap.AddCallerSkip(1)))
	globalLevel.Store(&level)
	globalFlush.Store(flush)
}

func SetGlobal(logger *zap.Logger, level zap.AtomicLevel, flush func() error) {
	if logger == nil {
		return
	}
	globalLogger.Store(logger)
	globalCallerLogger.Store(logger.WithOptions(zap.AddCallerSkip(1)))
	globalLevel.Store(&level)
	if flush != nil {
		globalFlush.Store(flush)
	}
}

func L() *zap.Logger {
	return globalLogger.Load()
}

func callerLogger() *zap.Logger {
	logger := globalCallerLogger.Load()
	if logger == nil {
		return L()
	}
	return logger
}

func SetLevel(level string) {
	if lvl := globalLevel.Load(); lvl != nil {
		_ = lvl.UnmarshalText([]byte(level))
	}
}

func GetLevel() Level {
	if lvl := globalLevel.Load(); lvl != nil {
		return lvl.Level()
	}
	return InfoLevel
}

func Debug(msg string, fields ...Field) { callerLogger().Debug(msg, fields...) }
func Info(msg string, fields ...Field)  { callerLogger().Info(msg, fields...) }
func Warn(msg string, fields ...Field)  { callerLogger().Warn(msg, fields...) }
func Error(msg string, fields ...Field) { callerLogger().Error(msg, fields...) }
func Fatal(msg string, fields ...Field) { callerLogger().Fatal(msg, fields...) }

func Debugf(format string, args ...any) { callerLogger().Sugar().Debugf(format, args...) }
func Infof(format string, args ...any)  { callerLogger().Sugar().Infof(format, args...) }
func Warnf(format string, args ...any)  { callerLogger().Sugar().Warnf(format, args...) }
func Errorf(format string, args ...any) { callerLogger().Sugar().Errorf(format, args...) }
func Fatalf(format string, args ...any) { callerLogger().Sugar().Fatalf(format, args...) }

func With(fields ...Field) *zap.Logger { return L().With(fields...) }

func WithContext(ctx context.Context) *zap.Logger {
	return WithContextLogger(L(), ctx)
}

func WithContextLogger(logger *zap.Logger, ctx context.Context) *zap.Logger {
	if logger == nil {
		return logger
	}
	if fields := extractContextFields(ctx); len(fields) > 0 {
		return logger.With(fields...)
	}
	return logger
}

func Named(name string) *zap.Logger { return L().Named(name) }
func Sync() error {
	if flush, ok := globalFlush.Load().(func() error); ok {
		_ = flush()
	}
	return L().Sync()
}

func Hooks(hooks ...func(zapcore.Entry) error) zap.Option {
	return zap.Hooks(hooks...)
}

func IncreaseLevel(lvl Level) zap.Option {
	return zap.IncreaseLevel(lvl)
}

func LevelHandler() http.Handler {
	if lvl := globalLevel.Load(); lvl != nil {
		return lvl
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "level control not supported", http.StatusNotImplemented)
	})
}

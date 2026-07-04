package log

import (
	"context"
	"errors"
	"os"
	"sync/atomic"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type state struct {
	logger *zap.Logger
	level  zap.AtomicLevel
	global *zap.Logger
	sugar  *zap.SugaredLogger
}

var current atomic.Pointer[state]

func init() {
	l, level, _ := New(defaultConfig())
	set(l, level)
	zap.ReplaceGlobals(l)
}

func set(l *zap.Logger, level zap.AtomicLevel) {
	global := l.WithOptions(zap.AddCallerSkip(1))
	current.Store(&state{
		logger: l,
		level:  level,
		global: global,
		sugar:  global.Sugar(),
	})
}

func load() *state {
	return current.Load()
}

func Init(cfg Config) (*zap.Logger, func(context.Context) error, error) {
	l, level, err := New(cfg)
	if err != nil {
		return nil, nil, err
	}

	set(l, level)
	zap.ReplaceGlobals(l)

	shutdown := func(context.Context) error {
		err := l.Sync()
		if isIgnorableSyncError(err) {
			return nil
		}
		return err
	}

	return l, shutdown, nil
}

// L returns the global logger for DI or direct use.
func L() *zap.Logger { return load().logger }

// SetLevel dynamically changes the global log level.
func SetLevel(l zapcore.Level) { load().level.SetLevel(l) }

// GetLevel returns the current log level.
func GetLevel() zapcore.Level { return load().level.Level() }

// AtomicLevel returns the underlying zap AtomicLevel for runtime level changes.
func AtomicLevel() zap.AtomicLevel { return load().level }

// With creates a child logger with additional fields.
func With(fields ...zap.Field) *zap.Logger { return load().logger.With(fields...) }

// Named creates a named child logger.
func Named(name string) *zap.Logger { return load().logger.Named(name) }

func Debug(msg string, fields ...zap.Field) { load().global.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { load().global.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { load().global.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { load().global.Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { load().global.Fatal(msg, fields...) }

func Debugf(template string, args ...any) { load().sugar.Debugf(template, args...) }
func Infof(template string, args ...any)  { load().sugar.Infof(template, args...) }
func Warnf(template string, args ...any)  { load().sugar.Warnf(template, args...) }
func Errorf(template string, args ...any) { load().sugar.Errorf(template, args...) }
func Fatalf(template string, args ...any) { load().sugar.Fatalf(template, args...) }

func isIgnorableSyncError(err error) bool {
	return errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY)
}

func New(cfg Config) (*zap.Logger, zap.AtomicLevel, error) {
	cfg.applyDefaults()

	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level.SetLevel(zapcore.InfoLevel)
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	if cfg.Development {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	}

	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	ws := zapcore.AddSync(os.Stdout)
	core := zapcore.NewCore(encoder, ws, level)

	options := []zap.Option{zap.ErrorOutput(zapcore.AddSync(os.Stderr))}
	if cfg.Caller {
		options = append(options, zap.AddCaller())
	}
	if cfg.Sampling != nil && cfg.Sampling.Enabled {
		options = append(options, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(
				core,
				time.Second,
				cfg.Sampling.Initial,
				cfg.Sampling.Thereafter,
			)
		}))
	}

	logger := zap.New(core, options...)
	return logger, level, nil
}

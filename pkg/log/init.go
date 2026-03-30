package log

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_logger *zap.Logger
	_level  zap.AtomicLevel
	_global *zap.Logger        // with CallerSkip(1) for package-level functions
	_sugar  *zap.SugaredLogger // from _global
)

func init() {
	l, level, _ := New(defaultConfig())
	set(l, level)
	zap.ReplaceGlobals(l)
}

func set(l *zap.Logger, level zap.AtomicLevel) {
	_logger = l
	_level = level
	_global = l.WithOptions(zap.AddCallerSkip(1))
	_sugar = _global.Sugar()
}

func Init(cfg Config) (*zap.Logger, func(context.Context) error, error) {
	l, level, err := New(cfg)
	if err != nil {
		return nil, nil, err
	}

	set(l, level)
	zap.ReplaceGlobals(l)

	shutdown := func(ctx context.Context) error {
		return l.Sync()
	}

	return l, shutdown, nil
}

// L returns the global logger for DI or direct use.
func L() *zap.Logger { return _logger }

// SetLevel dynamically changes the global log level.
func SetLevel(l zapcore.Level) { _level.SetLevel(l) }

// GetLevel returns the current log level.
func GetLevel() zapcore.Level { return _level.Level() }

// AtomicLevel returns the underlying AtomicLevel, which also implements
// http.Handler for GET/PUT log level changes at runtime.
func AtomicLevel() zap.AtomicLevel { return _level }

// With creates a child logger with additional fields.
func With(fields ...zap.Field) *zap.Logger { return _logger.With(fields...) }

// Named creates a named child logger.
func Named(name string) *zap.Logger { return _logger.Named(name) }

func Debug(msg string, fields ...zap.Field) { _global.Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)  { _global.Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { _global.Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { _global.Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { _global.Fatal(msg, fields...) }

func Debugf(template string, args ...any) { _sugar.Debugf(template, args...) }
func Infof(template string, args ...any)  { _sugar.Infof(template, args...) }
func Warnf(template string, args ...any)  { _sugar.Warnf(template, args...) }
func Errorf(template string, args ...any) { _sugar.Errorf(template, args...) }
func Fatalf(template string, args ...any) { _sugar.Fatalf(template, args...) }

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

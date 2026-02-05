package log

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg Config) (*zap.Logger, zap.AtomicLevel, func() error, error) {
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

	var (
		ws    zapcore.WriteSyncer
		flush func() error
	)

	if cfg.Async != nil && cfg.Async.Enabled {
		buffered := &zapcore.BufferedWriteSyncer{
			WS:            zapcore.AddSync(os.Stdout),
			Size:          cfg.Async.BufferSize,
			FlushInterval: time.Duration(cfg.Async.FlushInterval) * time.Millisecond,
		}
		ws = buffered
		flush = buffered.Stop
	} else {
		ws = zapcore.AddSync(os.Stdout)
		flush = func() error { return nil }
	}

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
	return logger, level, flush, nil
}

func NewDefault() (*zap.Logger, zap.AtomicLevel, func() error) {
	logger, level, flush, _ := New(DefaultConfig())
	return logger, level, flush
}

package bootstrap

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerParams struct {
	fx.In
	Level string `name:"log_level" optional:"true"`
	Env   string `name:"env" optional:"true"`
}

type LoggerResult struct {
	fx.Out
	Logger log.Logger
}

type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) Log(level log.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 || len(keyvals)%2 != 0 {
		l.logger.Warn("keyvals must be even number")
		return nil
	}

	fields := make([]zap.Field, 0, len(keyvals)/2)
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, zap.Any(key, keyvals[i+1]))
	}

	switch level {
	case log.LevelDebug:
		l.logger.Debug("", fields...)
	case log.LevelInfo:
		l.logger.Info("", fields...)
	case log.LevelWarn:
		l.logger.Warn("", fields...)
	case log.LevelError:
		l.logger.Error("", fields...)
	case log.LevelFatal:
		l.logger.Fatal("", fields...)
	}
	return nil
}

func NewLogger(params LoggerParams) (LoggerResult, error) {
	level := zapcore.InfoLevel
	if params.Level != "" {
		if err := level.UnmarshalText([]byte(params.Level)); err != nil {
			return LoggerResult{}, err
		}
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if params.Env == "production" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	zapLog := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	logger := &zapLogger{logger: zapLog}

	return LoggerResult{Logger: logger}, nil
}

func ProvideLogger(level, env string) fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func() string { return level },
				fx.ResultTags(`name:"log_level"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() string { return env },
				fx.ResultTags(`name:"env"`),
			),
		),
		fx.Provide(NewLogger),
	)
}

type fxLogger struct {
	l *zap.Logger
}

func (f *fxLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			f.l.Error("OnStart hook failed", zap.String("callee", e.FunctionName), zap.Error(e.Err))
		}
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			f.l.Error("OnStop hook failed", zap.String("callee", e.FunctionName), zap.Error(e.Err))
		}
	case *fxevent.Provided:
		if e.Err != nil {
			f.l.Error("provider failed", zap.Error(e.Err))
		}
	case *fxevent.Invoked:
		if e.Err != nil {
			f.l.Error("invoke failed", zap.String("function", e.FunctionName), zap.Error(e.Err))
		}
	case *fxevent.Started:
		if e.Err != nil {
			f.l.Error("start failed", zap.Error(e.Err))
		}
	case *fxevent.Stopped:
		if e.Err != nil {
			f.l.Error("stop failed", zap.Error(e.Err))
		}
	case *fxevent.RolledBack:
		if e.Err != nil {
			f.l.Error("rollback failed", zap.Error(e.Err))
		}
	}
}

// FxLogger returns an fx.Option that configures fx logging with zap
func FxLogger() fx.Option {
	return fx.WithLogger(func() fxevent.Logger {
		return &fxLogger{l: zap.L()}
	})
}

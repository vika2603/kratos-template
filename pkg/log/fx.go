package log

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Settings struct {
	Level       string
	Format      string
	Development bool
	Caller      bool
	Async       *AsyncConfig
}

type Result struct {
	fx.Out
	Logger   *zap.Logger
	Shutdown func(context.Context) error `name:"logger_shutdown"`
}

func NewFromSettings(lc fx.Lifecycle, settings Settings) (Result, error) {
	logger, shutdown, err := InitFromSettings(settings)
	if err != nil {
		return Result{}, err
	}

	lc.Append(fx.Hook{
		OnStop: shutdown,
	})

	return Result{
		Logger:   logger,
		Shutdown: shutdown,
	}, nil
}

func ProvideWithSettings(settings Settings) fx.Option {
	return fx.Provide(func(lc fx.Lifecycle) (Result, error) {
		return NewFromSettings(lc, settings)
	})
}

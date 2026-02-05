package log

import (
	"context"

	"go.uber.org/zap"
)

func InitFromSettings(settings Settings) (*zap.Logger, func(context.Context) error, error) {
	cfg := Config{
		Level:       settings.Level,
		Format:      settings.Format,
		Development: settings.Development,
		Caller:      settings.Caller,
		Async:       settings.Async,
	}

	logger, level, flush, err := New(cfg)
	if err != nil {
		return nil, nil, err
	}

	SetGlobal(logger, level, flush)

	shutdown := func(ctx context.Context) error {
		_ = flush()
		return logger.Sync()
	}

	return logger, shutdown, nil
}

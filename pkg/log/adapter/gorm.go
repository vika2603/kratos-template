package adapter

import (
	"context"
	"errors"
	"fmt"
	"kratos-template/pkg/log"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	defaultSlowThreshold = 200 * time.Millisecond
)

var _ gormlogger.Interface = (*GormAdapter)(nil)

type GormAdapter struct {
	logger        *zap.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

type GormOption func(*GormAdapter)

func WithSlowThreshold(d time.Duration) GormOption {
	return func(a *GormAdapter) { a.slowThreshold = d }
}

func WithGormLevel(level gormlogger.LogLevel) GormOption {
	return func(a *GormAdapter) { a.level = level }
}

func NewGormAdapter(logger *zap.Logger, opts ...GormOption) gormlogger.Interface {
	a := &GormAdapter{
		logger:        logger.WithOptions(zap.AddCallerSkip(1)).Named("gorm"),
		level:         gormlogger.Warn,
		slowThreshold: defaultSlowThreshold,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *GormAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &GormAdapter{
		logger:        a.logger,
		level:         level,
		slowThreshold: a.slowThreshold,
	}
}

func (a *GormAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	if a.level >= gormlogger.Info {
		log.WithContextLogger(ctx, a.logger).Sugar().Infof(msg, data...)
	}
}

func (a *GormAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	if a.level >= gormlogger.Warn {
		log.WithContextLogger(ctx, a.logger).Sugar().Warnf(msg, data...)
	}
}

func (a *GormAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	if a.level >= gormlogger.Error {
		log.WithContextLogger(ctx, a.logger).Sugar().Errorf(msg, data...)
	}
}

func (a *GormAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if a.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	l := log.WithContextLogger(ctx, a.logger).With(
		zap.Duration("elapsed", elapsed),
		zap.String("sql", sql),
		zap.Int64("rows", rows),
	)

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		l.Error("query error", zap.NamedError("err", err))
	case elapsed > a.slowThreshold && a.slowThreshold > 0:
		l.Warn(fmt.Sprintf("slow query >= %v", a.slowThreshold))
	case a.level >= gormlogger.Info:
		l.Debug("query executed")
	}
}

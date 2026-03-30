package adapter

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.uber.org/zap"

	"kratos-template/pkg/log"
)

var _ hlog.FullLogger = (*HertzAdapter)(nil)

type HertzAdapter struct {
	logger *zap.Logger
	level  hlog.Level
}

func NewHertzAdapter() *HertzAdapter {
	return &HertzAdapter{
		logger: log.L().WithOptions(zap.AddCallerSkip(1)).Named("hertz"),
		level:  hlog.LevelInfo,
	}
}

func (a *HertzAdapter) Trace(v ...interface{}) {
	a.logger.Debug(fmt.Sprint(v...))
}

func (a *HertzAdapter) Debug(v ...interface{}) {
	a.logger.Debug(fmt.Sprint(v...))
}

func (a *HertzAdapter) Info(v ...interface{}) {
	a.logger.Info(fmt.Sprint(v...))
}

func (a *HertzAdapter) Notice(v ...interface{}) {
	a.logger.Info(fmt.Sprint(v...))
}

func (a *HertzAdapter) Warn(v ...interface{}) {
	a.logger.Warn(fmt.Sprint(v...))
}

func (a *HertzAdapter) Error(v ...interface{}) {
	a.logger.Error(fmt.Sprint(v...))
}

func (a *HertzAdapter) Fatal(v ...interface{}) {
	a.logger.Fatal(fmt.Sprint(v...))
}

func (a *HertzAdapter) Tracef(format string, v ...interface{}) {
	a.logger.Sugar().Debugf(format, v...)
}

func (a *HertzAdapter) Debugf(format string, v ...interface{}) {
	a.logger.Sugar().Debugf(format, v...)
}

func (a *HertzAdapter) Infof(format string, v ...interface{}) {
	a.logger.Sugar().Infof(format, v...)
}

func (a *HertzAdapter) Noticef(format string, v ...interface{}) {
	a.logger.Sugar().Infof(format, v...)
}

func (a *HertzAdapter) Warnf(format string, v ...interface{}) {
	a.logger.Sugar().Warnf(format, v...)
}

func (a *HertzAdapter) Errorf(format string, v ...interface{}) {
	a.logger.Sugar().Errorf(format, v...)
}

func (a *HertzAdapter) Fatalf(format string, v ...interface{}) {
	a.logger.Sugar().Fatalf(format, v...)
}

func (a *HertzAdapter) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	log.WithContextLogger(a.logger, ctx).Sugar().Debugf(format, v...)
}

func (a *HertzAdapter) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	log.WithContextLogger(a.logger, ctx).Sugar().Debugf(format, v...)
}

func (a *HertzAdapter) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	log.WithContextLogger(a.logger, ctx).Sugar().Infof(format, v...)
}

func (a *HertzAdapter) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	log.WithContextLogger(a.logger, ctx).Sugar().Infof(format, v...)
}

func (a *HertzAdapter) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	log.WithContextLogger(a.logger, ctx).Sugar().Warnf(format, v...)
}

func (a *HertzAdapter) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	log.WithContextLogger(a.logger, ctx).Sugar().Errorf(format, v...)
}

func (a *HertzAdapter) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	log.WithContextLogger(a.logger, ctx).Sugar().Fatalf(format, v...)
}

func (a *HertzAdapter) SetLevel(level hlog.Level) {
	a.level = level
}

func (a *HertzAdapter) SetOutput(_ io.Writer) {}

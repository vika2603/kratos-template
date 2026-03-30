package adapter

import (
	"strings"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"kratos-template/pkg/log"
)

var _ fxevent.Logger = (*FxAdapter)(nil)

type FxAdapter struct{}

func NewFxAdapter() fxevent.Logger {
	return &FxAdapter{}
}

func (a *FxAdapter) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		log.Debug("fx: OnStart hook executing",
			zap.String("callee", e.FunctionName),
			zap.String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			log.Error("fx: OnStart hook failed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Error(e.Err),
			)
		} else {
			log.Debug("fx: OnStart hook executed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Duration("runtime", e.Runtime),
			)
		}
	case *fxevent.OnStopExecuting:
		log.Debug("fx: OnStop hook executing",
			zap.String("callee", e.FunctionName),
			zap.String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			log.Error("fx: OnStop hook failed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Error(e.Err),
			)
		} else {
			log.Debug("fx: OnStop hook executed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.Duration("runtime", e.Runtime),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			log.Error("fx: supply failed",
				zap.String("type", e.TypeName),
				zap.Error(e.Err),
			)
		}
	case *fxevent.Provided:
		if e.Err != nil {
			log.Error("fx: provide failed",
				zap.String("constructor", e.ConstructorName),
				zap.Error(e.Err),
			)
		}
	case *fxevent.Decorated:
		if e.Err != nil {
			log.Error("fx: decorate failed",
				zap.String("decorator", e.DecoratorName),
				zap.Error(e.Err),
			)
		}
	case *fxevent.Invoked:
		if e.Err != nil {
			log.Error("fx: invoke failed",
				zap.String("function", e.FunctionName),
				zap.Error(e.Err),
				zap.String("stack", e.Trace),
			)
		}
	case *fxevent.Stopping:
		log.Info("fx: stopping",
			zap.String("signal", strings.ToUpper(e.Signal.String())),
		)
	case *fxevent.Stopped:
		if e.Err != nil {
			log.Error("fx: stop failed", zap.Error(e.Err))
		}
	case *fxevent.RollingBack:
		log.Error("fx: rolling back", zap.Error(e.StartErr))
	case *fxevent.RolledBack:
		if e.Err != nil {
			log.Error("fx: rollback failed", zap.Error(e.Err))
		}
	case *fxevent.Started:
		if e.Err != nil {
			log.Error("fx: start failed", zap.Error(e.Err))
		} else {
			log.Info("fx: started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			log.Error("fx: logger initialization failed", zap.Error(e.Err))
		}
	}
}

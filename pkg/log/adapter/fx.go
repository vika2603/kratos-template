package adapter

import (
	"strings"

	"go.uber.org/fx/fxevent"

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
			log.String("callee", e.FunctionName),
			log.String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			log.Error("fx: OnStart hook failed",
				log.String("callee", e.FunctionName),
				log.String("caller", e.CallerName),
				log.Err(e.Err),
			)
		} else {
			log.Debug("fx: OnStart hook executed",
				log.String("callee", e.FunctionName),
				log.String("caller", e.CallerName),
				log.Duration("runtime", e.Runtime),
			)
		}
	case *fxevent.OnStopExecuting:
		log.Debug("fx: OnStop hook executing",
			log.String("callee", e.FunctionName),
			log.String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			log.Error("fx: OnStop hook failed",
				log.String("callee", e.FunctionName),
				log.String("caller", e.CallerName),
				log.Err(e.Err),
			)
		} else {
			log.Debug("fx: OnStop hook executed",
				log.String("callee", e.FunctionName),
				log.String("caller", e.CallerName),
				log.Duration("runtime", e.Runtime),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			log.Error("fx: supply failed",
				log.String("type", e.TypeName),
				log.Err(e.Err),
			)
		}
	case *fxevent.Provided:
		if e.Err != nil {
			log.Error("fx: provide failed",
				log.String("constructor", e.ConstructorName),
				log.Err(e.Err),
			)
		}
	case *fxevent.Decorated:
		if e.Err != nil {
			log.Error("fx: decorate failed",
				log.String("decorator", e.DecoratorName),
				log.Err(e.Err),
			)
		}
	case *fxevent.Invoked:
		if e.Err != nil {
			log.Error("fx: invoke failed",
				log.String("function", e.FunctionName),
				log.Err(e.Err),
				log.String("stack", e.Trace),
			)
		}
	case *fxevent.Stopping:
		log.Info("fx: stopping",
			log.String("signal", strings.ToUpper(e.Signal.String())),
		)
	case *fxevent.Stopped:
		if e.Err != nil {
			log.Error("fx: stop failed", log.Err(e.Err))
		}
	case *fxevent.RollingBack:
		log.Error("fx: rolling back", log.Err(e.StartErr))
	case *fxevent.RolledBack:
		if e.Err != nil {
			log.Error("fx: rollback failed", log.Err(e.Err))
		}
	case *fxevent.Started:
		if e.Err != nil {
			log.Error("fx: start failed", log.Err(e.Err))
		} else {
			log.Info("fx: started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			log.Error("fx: logger initialization failed", log.Err(e.Err))
		}
	}
}

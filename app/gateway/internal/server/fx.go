package server

import (
	"github.com/go-kratos/kratos/v2/transport"
	"go.uber.org/fx"
)

var Module = fx.Module("server",
	fx.Provide(NewHertzServer),
	fx.Provide(
		fx.Annotate(
			func(s *HertzServer) transport.Server { return s },
			fx.ResultTags(`group:"servers"`),
		),
	),
)

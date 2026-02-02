package server

import (
	"github.com/go-kratos/kratos/v2/transport"
	"go.uber.org/fx"
)

var Module = fx.Module("server",
	fx.Provide(
		fx.Annotate(
			NewHTTPServer,
			fx.As(new(transport.Server)),
			fx.ResultTags(`group:"servers"`),
		),
	),
	fx.Provide(
		fx.Annotate(
			NewGRPCServer,
			fx.As(new(transport.Server)),
			fx.ResultTags(`group:"servers"`),
		),
	),
)

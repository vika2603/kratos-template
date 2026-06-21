package server

import (
	"github.com/go-kratos/kratos/v2/transport"
	"go.uber.org/fx"
)

var Module = fx.Module("user.server",
	fx.Provide(
		fx.Annotate(
			NewGRPCServer,
			fx.As(new(transport.Server)),
			fx.ResultTags(`group:"servers"`),
		),
	),
)

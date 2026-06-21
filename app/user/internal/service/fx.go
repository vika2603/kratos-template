package service

import "go.uber.org/fx"

var Module = fx.Module("user.service",
	fx.Provide(NewUserService),
)

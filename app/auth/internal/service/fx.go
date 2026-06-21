package service

import "go.uber.org/fx"

var Module = fx.Module("auth.service",
	fx.Provide(NewAuthService),
)

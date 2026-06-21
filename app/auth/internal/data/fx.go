package data

import (
	"context"

	"go.uber.org/fx"
)

var Module = fx.Module("auth.data",
	fx.Provide(NewDB),
	fx.Provide(NewData),
	fx.Provide(NewAuthUserRepo),
	fx.Invoke(registerLifecycle),
)

func registerLifecycle(lc fx.Lifecycle, cleanup func()) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cleanup()
			return nil
		},
	})
}

package gateway

import (
	"context"
	"flag"

	"go.uber.org/fx"

	"kratos-template/app/gateway/biz/router"
	"kratos-template/app/gateway/internal/client"
	"kratos-template/app/gateway/internal/conf"
	"kratos-template/app/gateway/internal/server"
	"kratos-template/app/gateway/pkg/handler"
	authmw "kratos-template/app/gateway/pkg/middleware/auth"
	pkgauth "kratos-template/pkg/auth"
	"kratos-template/pkg/bootstrap"
)

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/gateway.yaml", "config path")
}

func Run() {
	flag.Parse()
	bootstrap.Run[conf.Bootstrap](flagConf, "config/gateway/",
		client.ProvideClients(),

		fx.Provide(server.NewHertzServer, handler.Default),

		fx.Provide(provideAuthConfig),
		pkgauth.Module,
		fx.Invoke(func(mgr *pkgauth.JWTManager) {
			authmw.Init(mgr)
		}),

		router.Options(),

		fx.Invoke(func(lc fx.Lifecycle, s *server.HertzServer) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error { return s.Start(ctx) },
				OnStop:  func(ctx context.Context) error { return s.Stop(ctx) },
			})
		}),
	)
}

func provideAuthConfig(cfg *conf.Bootstrap) pkgauth.Config {
	return pkgauth.Config{
		Secret:      cfg.Auth.JwtSecret,
		ExpiryHours: 24,
	}
}

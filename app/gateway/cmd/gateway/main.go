package main

import (
	"context"
	"os"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/fx"

	"kratos-template/app/gateway/biz/module"
	"kratos-template/app/gateway/internal/client"
	"kratos-template/app/gateway/internal/conf"
	"kratos-template/app/gateway/internal/server"
	"kratos-template/app/gateway/pkg/auth"
	"kratos-template/app/gateway/pkg/handler"
	"kratos-template/pkg/bootstrap"
)

func main() {
	bc := loadConfig()

	fx.New(
		fx.Supply(bc),
		bootstrap.ProvideLogger(bc.Log.Level, bc.Log.Env),
		bootstrap.ProvideRegistry(bc.Registry.Consul.Address, bc.Registry.Consul.Scheme),

		fx.Provide(
			client.NewAuthClient,
			client.NewUserClient,
			client.NewAssetClient,
		),

		fx.Provide(
			server.NewHertzServer,
			handler.Default,
		),

		module.Modules(),
		fx.Invoke(func(lc fx.Lifecycle, s *server.HertzServer) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error { return s.Start(ctx) },
				OnStop:  func(ctx context.Context) error { return s.Stop(ctx) },
			})
		}),
	).Run()
}

func loadConfig() *conf.Bootstrap {
	c := config.New(config.WithSource(file.NewSource("configs/config.yaml")))
	if err := c.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		log.Fatalf("failed to scan config: %v", err)
	}

	if addr := os.Getenv("CONSUL_ADDR"); addr != "" {
		bc.Registry.Consul.Address = addr
	}

	auth.Init(bc.Auth.JwtSecret)

	return &bc
}

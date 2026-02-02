package config

import (
	"github.com/go-kratos/kratos/v2/config"
	"go.uber.org/fx"

	"kratos-template/app/gateway/internal/conf"
)

func LoadBootstrap(cfg config.Config) (*conf.Bootstrap, error) {
	var bc conf.Bootstrap
	if err := cfg.Scan(&bc); err != nil {
		return nil, err
	}
	return &bc, nil
}

func Module() fx.Option {
	return fx.Provide(LoadBootstrap)
}

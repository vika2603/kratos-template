package bootstrap

import (
	"context"
	"os"

	"github.com/go-kratos/kratos/contrib/config/consul/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/hashicorp/consul/api"
	"go.uber.org/fx"
)

func ProvideConfig(localPath string) fx.Option {
	return ProvideConfigWithConsul(localPath, "", "")
}

func ProvideConfigWithConsul(localPath, consulPath, consulAddr string) fx.Option {
	return fx.Options(
		fx.Provide(func() (config.Config, error) {
			return newConfig(localPath, consulPath, consulAddr)
		}),
		fx.Invoke(func(lc fx.Lifecycle, c config.Config) {
			lc.Append(fx.Hook{
				OnStop: func(context.Context) error {
					return c.Close()
				},
			})
		}),
	)
}

func newConfig(localPath, consulPath, consulAddr string) (config.Config, error) {
	if consulAddr == "" {
		consulAddr = os.Getenv("CONSUL_ADDR")
	}

	consulConfigPath := os.Getenv("CONSUL_CONFIG_PATH")
	if consulConfigPath != "" {
		consulPath = consulConfigPath
	}

	var sources []config.Source

	if consulAddr != "" && consulPath != "" {
		consulClient, err := api.NewClient(&api.Config{
			Address: consulAddr,
		})
		if err == nil {
			cs, err := consul.New(consulClient, consul.WithPath(consulPath))
			if err == nil {
				sources = append(sources, cs)
				log.Infof("Config: using Consul source %s%s", consulAddr, consulPath)
			}
		}
	}

	if localPath != "" {
		if _, err := os.Stat(localPath); err == nil {
			sources = append(sources, file.NewSource(localPath))
			log.Infof("Config: using local file %s", localPath)
		}
	}

	c := config.New(config.WithSource(sources...))
	if err := c.Load(); err != nil {
		return nil, err
	}

	return c, nil
}

func WatchConfig[T any](c config.Config, key string, callback func(*T)) error {
	return c.Watch(key, func(k string, v config.Value) {
		var cfg T
		if err := v.Scan(&cfg); err != nil {
			return
		}
		callback(&cfg)
	})
}

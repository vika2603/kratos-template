package bootstrap

import (
	"cmp"
	"os"

	"github.com/go-kratos/kratos/contrib/config/consul/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	kratoslog "github.com/go-kratos/kratos/v2/log"
	"github.com/hashicorp/consul/api"

	"kratos-template/pkg/conf"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
)

func init() {
	kratoslog.SetLogger(adapter.NewKratosGlobalAdapter())
}

const consulAddrEnv = "CONSUL_ADDR"

func NewConfig(localPath, consulPath, consulAddr string) (config.Config, error) {
	consulAddr = cmp.Or(consulAddr, os.Getenv(consulAddrEnv))
	consulPath = cmp.Or(os.Getenv("CONSUL_CONFIG_PATH"), consulPath)

	var sources []config.Source

	if localPath != "" {
		if _, err := os.Stat(localPath); err == nil {
			sources = append(sources, file.NewSource(localPath))
			log.Infof("Config: using local file %s", localPath)
		}
	}

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

	c := config.New(config.WithSource(sources...))
	if err := c.Load(); err != nil {
		return nil, err
	}

	return c, nil
}

func LoadConfig[T any](cfg config.Config) (*T, error) {
	var bc T
	if err := cfg.Scan(&bc); err != nil {
		return nil, err
	}
	return &bc, nil
}

func ScanCommonConfig(cfg config.Config) (*conf.CommonConfig, error) {
	var cc conf.CommonConfig
	if err := cfg.Scan(&cc); err != nil {
		return nil, err
	}
	if name := os.Getenv("SERVICE_NAME"); name != "" {
		if cc.Service == nil {
			cc.Service = &conf.Service{}
		}
		cc.Service.Name = name
	}
	if version := os.Getenv("SERVICE_VERSION"); version != "" {
		if cc.Service == nil {
			cc.Service = &conf.Service{}
		}
		cc.Service.Version = version
	}
	return &cc, nil
}

func hostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
}

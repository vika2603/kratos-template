package bootstrap

import (
	"cmp"
	"errors"
	"fmt"
	"kratos-template/pkg/log"
	"kratos-template/pkg/log/adapter"
	"os"

	"github.com/go-kratos/kratos/contrib/config/consul/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	kratoslog "github.com/go-kratos/kratos/v2/log"
	"github.com/hashicorp/consul/api"
)

func init() {
	kratoslog.SetLogger(adapter.NewKratosGlobalAdapter())
}

const consulAddrEnv = "CONSUL_ADDR"

// NewConfig assembles config from layered sources and loads it. Both the local
// file and Consul may apply at once — see configSources for precedence.
func NewConfig(localPath, consulPath string) (config.Config, error) {
	sources, err := configSources(localPath, consulPath)
	if err != nil {
		return nil, err
	}

	c := config.New(config.WithSource(sources...))
	if err := c.Load(); err != nil {
		return nil, err
	}
	return c, nil
}

// configSources layers sources low→high precedence: local file (committed
// defaults) < Consul (deployed overrides), with env vars overriding individual
// values later at their read sites. Both are optional; at least one must resolve.
// Consul is only added when CONSUL_ADDR is set, and a configured-but-unreachable
// Consul is a hard error rather than a silent fallthrough.
func configSources(localPath, consulPath string) ([]config.Source, error) {
	var sources []config.Source

	if localPath != "" {
		if _, err := os.Stat(localPath); err == nil {
			sources = append(sources, file.NewSource(localPath))
			log.Infof("Config: file %s", localPath)
		}
	}

	if addr := os.Getenv(consulAddrEnv); addr != "" {
		path := cmp.Or(os.Getenv("CONSUL_CONFIG_PATH"), consulPath)
		cs, err := consulSource(addr, path)
		if err != nil {
			return nil, fmt.Errorf("consul %s%s: %w", addr, path, err)
		}
		sources = append(sources, cs)
		log.Infof("Config: Consul %s%s", addr, path)
	}

	if len(sources) == 0 {
		return nil, errors.New("no config source: pass -conf <file> or set CONSUL_ADDR")
	}
	return sources, nil
}

func consulSource(addr, path string) (config.Source, error) {
	client, err := api.NewClient(&api.Config{Address: addr})
	if err != nil {
		return nil, err
	}
	return consul.New(client, consul.WithPath(path))
}

func LoadConfig[T any](cfg config.Config) (*T, error) {
	var bc T
	if err := cfg.Scan(&bc); err != nil {
		return nil, err
	}
	return &bc, nil
}

func hostname() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}
	return hostname
}

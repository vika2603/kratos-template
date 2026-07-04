package registry

import (
	"kratos-template/pkg/log"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
	"go.uber.org/fx"
)

type Params struct {
	fx.In
	Address string `name:"consul_address" optional:"true"`
	Scheme  string `name:"consul_scheme" optional:"true"`
}

type Result struct {
	fx.Out
	Registry  registry.Registrar
	Discovery registry.Discovery
}

func New(params Params) (Result, error) {
	addr := params.Address

	// No Consul configured (e.g. local single-service run) — skip registration.
	// Consumers must treat Registrar/Discovery as optional.
	if addr == "" {
		log.Info("registry: consul address not set, service discovery disabled")
		return Result{}, nil
	}

	cfg := api.DefaultConfig()
	cfg.Address = addr
	if params.Scheme != "" {
		cfg.Scheme = params.Scheme
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return Result{}, err
	}

	reg := consul.New(client,
		consul.WithHealthCheck(true),
		consul.WithHealthCheckInterval(10),
		consul.WithDeregisterCriticalServiceAfter(60),
	)
	return Result{
		Registry:  reg,
		Discovery: reg,
	}, nil
}

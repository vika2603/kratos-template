package registry

import (
	"os"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
	"go.uber.org/fx"
)

type Settings struct {
	Address string
	Scheme  string
}

type Params struct {
	fx.In
	Address string `name:"consul_address"`
	Scheme  string `name:"consul_scheme" optional:"true"`
}

type Result struct {
	fx.Out
	Registry  registry.Registrar
	Discovery registry.Discovery
}

func New(params Params) (Result, error) {
	cfg := api.DefaultConfig()
	addr := os.Getenv("CONSUL_ADDR")
	if addr == "" {
		addr = params.Address
	}
	cfg.Address = addr
	if params.Scheme != "" {
		cfg.Scheme = params.Scheme
	}

	client, err := api.NewClient(cfg)
	if err != nil {
		return Result{}, err
	}

	reg := consul.New(client)
	return Result{
		Registry:  reg,
		Discovery: reg,
	}, nil
}

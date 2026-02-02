package bootstrap

import (
	"os"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
	"go.uber.org/fx"
)

type RegistryParams struct {
	fx.In
	Address string `name:"consul_address"`
	Scheme  string `name:"consul_scheme" optional:"true"`
}

type RegistryResult struct {
	fx.Out
	Registry  registry.Registrar
	Discovery registry.Discovery
}

func NewRegistry(params RegistryParams) (RegistryResult, error) {
	config := api.DefaultConfig()
	addr := os.Getenv("CONSUL_ADDR")
	if addr == "" {
		addr = params.Address
	}
	config.Address = addr
	if params.Scheme != "" {
		config.Scheme = params.Scheme
	}

	client, err := api.NewClient(config)
	if err != nil {
		return RegistryResult{}, err
	}

	reg := consul.New(client)
	return RegistryResult{
		Registry:  reg,
		Discovery: reg,
	}, nil
}

func ProvideRegistry(address, scheme string) fx.Option {
	return fx.Options(
		fx.Provide(
			fx.Annotate(
				func() string { return address },
				fx.ResultTags(`name:"consul_address"`),
			),
		),
		fx.Provide(
			fx.Annotate(
				func() string { return scheme },
				fx.ResultTags(`name:"consul_scheme"`),
			),
		),
		fx.Provide(NewRegistry),
	)
}

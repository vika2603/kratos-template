package module

import "go.uber.org/fx"

var modules []fx.Option

func register(opt fx.Option) {
	modules = append(modules, opt)
}

func Modules() fx.Option {
	return fx.Options(modules...)
}

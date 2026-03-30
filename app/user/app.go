package user

import (
	"flag"

	"kratos-template/app/user/internal/biz"
	"kratos-template/app/user/internal/conf"
	"kratos-template/app/user/internal/data"
	"kratos-template/app/user/internal/server"
	"kratos-template/app/user/internal/service"
	"kratos-template/pkg/bootstrap"
)

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/user.yaml", "config path")
}

func Run() {
	flag.Parse()
	bootstrap.Run[conf.Bootstrap](flagConf, "config/user/",
		bootstrap.WithKratosApp(),
		data.Module,
		biz.Module,
		service.Module,
		server.Module,
	)
}

package main

import (
	"flag"

	"kratos-template/app/auth/internal/biz"
	"kratos-template/app/auth/internal/conf"
	"kratos-template/app/auth/internal/data"
	"kratos-template/app/auth/internal/server"
	"kratos-template/app/auth/internal/service"
	"kratos-template/pkg/bootstrap"
)

var flagConf string

func init() {
	flag.StringVar(&flagConf, "conf", "configs/auth.yaml", "config path")
}

func main() {
	flag.Parse()
	bootstrap.Run[conf.Bootstrap](flagConf, "config/auth/",
		bootstrap.WithKratosApp(),
		data.Module,
		biz.Module,
		service.Module,
		server.Module,
	)
}

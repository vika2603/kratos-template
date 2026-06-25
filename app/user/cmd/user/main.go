package main

import (
	"kratos-template/app/user/internal/biz"
	"kratos-template/app/user/internal/conf"
	"kratos-template/app/user/internal/data"
	"kratos-template/app/user/internal/server"
	"kratos-template/app/user/internal/service"
	"kratos-template/pkg/bootstrap"
)

func main() {
	bootstrap.Run[conf.Bootstrap]("user",
		bootstrap.WithKratosApp(),
		data.Module,
		biz.Module,
		service.Module,
		server.Module,
	)
}

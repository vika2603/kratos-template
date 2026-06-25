package main

import (
	"kratos-template/app/auth/internal/biz"
	"kratos-template/app/auth/internal/conf"
	"kratos-template/app/auth/internal/data"
	"kratos-template/app/auth/internal/server"
	"kratos-template/app/auth/internal/service"
	"kratos-template/pkg/bootstrap"
)

func main() {
	bootstrap.Run[conf.Bootstrap]("auth",
		bootstrap.WithKratosApp(),
		data.Module,
		biz.Module,
		service.Module,
		server.Module,
	)
}
